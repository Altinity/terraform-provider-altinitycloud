package env

import (
	"context"
	"errors"
	"fmt"
	"time"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/auth"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// StatusCheckFunc checks if the env is still being deleted.
// Returns (pendingDelete bool, err error).
// err should be the raw SDK error (not-found handling is done by the caller),
// or ErrEnvNotFound when the query succeeded but returned no environment.
type StatusCheckFunc func(ctx context.Context, name string) (bool, error)

// ErrEnvNotFound reports an env that is already gone. A status query can succeed
// and still return a null env, which carries no SDK error for IsNotFoundError to
// recognize; returning (false, nil) instead would strand a pendingMFA delete in
// PENDING_MFA until the MFA timeout, so callers return this to mean "deleted".
var ErrEnvNotFound = errors.New("environment not found")

var MFATimeout = 5 * time.Minute
var DeleteTimeout = 60 * time.Minute
var DeletePollInterval = 30 * time.Second

type EnvResourceBase struct {
	Client *client.Client
	Auth   *auth.Auth
}

func (r *EnvResourceBase) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*sdk.AltinityCloudSDK)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.AltinityCloudSDK, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.Client = sdk.Client
	r.Auth = sdk.Auth
}

func (r *EnvResourceBase) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	diags := resp.State.SetAttribute(ctx, path.Root("force_destroy"), false)
	resp.Diagnostics.Append(diags...)
	diags = resp.State.SetAttribute(ctx, path.Root("force_destroy_clusters"), false)
	resp.Diagnostics.Append(diags...)
	diags = resp.State.SetAttribute(ctx, path.Root("skip_deprovision_on_destroy"), false)
	resp.Diagnostics.Append(diags...)
	diags = resp.State.SetAttribute(ctx, path.Root("allow_delete_while_disconnected"), false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *EnvResourceBase) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		var skipDeprovision types.Bool
		req.State.GetAttribute(ctx, path.Root("skip_deprovision_on_destroy"), &skipDeprovision)

		if skipDeprovision.ValueBool() {
			resp.Diagnostics.AddAttributeWarning(path.Root("skip_deprovision_on_destroy"), "Skip Deprovision on Destroy", "This resource is using the 'skip_deprovision_on_destroy'.\nUse this with precaution as it will delete the environment without deleting any of your cloud resources.")
		}
		return
	}

	// spec_revision is server-reassigned each update; unknown-on-change so the plan doesn't pin the stale value.
	if !req.State.Raw.IsNull() && !req.Plan.Raw.Equal(req.State.Raw) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("spec_revision"), types.Int64Unknown())...)
	}
}

func WaitForDeletion(ctx context.Context, resp *resource.DeleteResponse, envName string, pendingMfa bool, checkStatus StatusCheckFunc, deleteTimeout time.Duration, mfaTimeout time.Duration) {
	waitForDeletion(ctx, resp, envName, pendingMfa, checkStatus, deleteTimeout, mfaTimeout, DeletePollInterval)
}

// pollInterval is a parameter so tests can shorten it without mutating the
// package-level default, which races when they run in parallel.
func waitForDeletion(ctx context.Context, resp *resource.DeleteResponse, envName string, pendingMfa bool, checkStatus StatusCheckFunc, deleteTimeout time.Duration, mfaTimeout time.Duration, pollInterval time.Duration) {
	if deleteTimeout == 0 {
		deleteTimeout = DeleteTimeout
	}
	if mfaTimeout == 0 {
		mfaTimeout = MFATimeout
	}

	mfaStart := time.Now()
	stateConf := &retry.StateChangeConf{
		Pending: []string{"PENDING_MFA", "DELETING"},
		Target:  []string{"DELETED"},
		Refresh: func() (interface{}, string, error) {
			pendingDelete, err := checkStatus(ctx, envName)
			if err != nil {
				notFound, _ := client.IsNotFoundError(err)
				if notFound || errors.Is(err, ErrEnvNotFound) {
					tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
					return envName, "DELETED", nil
				}
				// StateChangeConf aborts on any non-nil error from Refresh. Treat other
				// errors (e.g. transient 5xx/network) as still deleting so polling
				// continues until deleteTimeout or the resource is gone.
				tflog.Trace(ctx, "error while polling deletion status; will retry", map[string]interface{}{
					"name":  envName,
					"error": err.Error(),
				})
				return envName, "DELETING", nil
			}

			if !pendingDelete {
				if !pendingMfa {
					tflog.Trace(ctx, "deleted resource (pendingDelete cleared)", map[string]interface{}{"name": envName})
					return envName, "DELETED", nil
				}
				if time.Since(mfaStart) > mfaTimeout {
					return nil, "", fmt.Errorf("timeout reached while waiting for MFA to be confirmed.\nPlease check your MFA device, confirm deletion and run `terraform destroy` again")
				}
				return envName, "PENDING_MFA", nil
			}

			return envName, "DELETING", nil
		},
		Timeout:      deleteTimeout,
		PollInterval: pollInterval,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		clientsupport.AddSupportError(&resp.Diagnostics, "Delete Error", fmt.Sprintf("Error waiting for env %s to be deleted: %s", envName, err))
	}
}
