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
)

var MFA_TIMEOUT = time.Duration(5) * time.Minute
var DELETE_TIMEOUT = time.Duration(60) * time.Minute
var DELETE_POLL_INTERVAL = 30 * time.Second

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
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
