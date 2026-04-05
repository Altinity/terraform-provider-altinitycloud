package env

import (
	"context"
	"fmt"
	"time"

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
// err should be the raw SDK error (not-found handling is done by the caller).
type StatusCheckFunc func(ctx context.Context, name string) (bool, error)

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
	}
}

func WaitForDeletion(ctx context.Context, resp *resource.DeleteResponse, envName string, pendingMfa bool, checkStatus StatusCheckFunc, deleteTimeout time.Duration, mfaTimeout time.Duration) {
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
				if notFound {
					tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
					return envName, "DELETED", nil
				}
				return nil, "", err
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
		PollInterval: DeletePollInterval,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("Error waiting for env %s to be deleted: %s", envName, err))
	}
}
