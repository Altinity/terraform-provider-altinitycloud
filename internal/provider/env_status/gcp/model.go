package env_status

import (
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GCPEnvStatusModel struct {
	Id                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	WaitForAppliedSpecRevision types.Int64  `tfsdk:"wait_for_applied_spec_revision"`
	AppliedSpecRevision        types.Int64  `tfsdk:"applied_spec_revision"`
	PendingDelete              types.Bool   `tfsdk:"pending_delete"`
}

func (model *GCPEnvStatusModel) toModel(env sdk.GetGCPEnvStatus_GCPEnv) {
	model.Name = types.StringValue(env.Name)
	model.AppliedSpecRevision = types.Int64Value(env.Status.AppliedSpecRevision)
	model.PendingDelete = types.BoolValue(env.Status.PendingDelete)
}
