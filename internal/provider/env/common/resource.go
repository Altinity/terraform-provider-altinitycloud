package env

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var MFA_TIMEOUT = time.Duration(5) * time.Minute
var DELETE_TIMEOUT = time.Duration(60) * time.Minute
var DELETE_POLL_INTERVAL = 30 * time.Second

func Import(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		var skipDeprovision types.Bool
		req.State.GetAttribute(ctx, path.Root("skip_deprovision_on_destroy"), &skipDeprovision)

		if skipDeprovision.ValueBool() {
			resp.Diagnostics.AddAttributeWarning(path.Root("skip_deprovision_on_destroy"), "Skip Deprovision on Destroy", "This resource is using the 'skip_deprovision_on_destroy'.\nUse this with precaution as it will delete the environment without deleting any of your cloud resources.")
		}
	}
}
