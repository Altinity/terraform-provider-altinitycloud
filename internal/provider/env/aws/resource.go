package env

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &AWSEnvResource{}
var _ resource.ResourceWithImportState = &AWSEnvResource{}

func NewAWSEnvResource() resource.Resource {
	return &AWSEnvResource{}
}

type AWSEnvResource struct {
	common.EnvResourceBase
}

func (r *AWSEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_aws"
}

func (r *AWSEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AWSEnvResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "creating resource", map[string]interface{}{"name": envName})

	sdkEnv, _, diags := data.toSDK(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiResp, err := r.Client.CreateAWSEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	// Reorder node groups, zones and tags to respect order in the user's configuration
	apiResp.CreateAWSEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.CreateAWSEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AWSEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.CreateAWSEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.CreateAWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.CreateAWSEnv.Spec.Tags = common.ReorderByKey(data.Tags, apiResp.CreateAWSEnv.Spec.Tags,
		func(m common.KeyValueModel) string { return m.Key.ValueString() },
		func(s *client.AWSEnvSpecFragment_Tags) string { return s.Key },
	)
	apiResp.CreateAWSEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.CreateAWSEnv.Spec.PeeringConnections,
		func(m AWSEnvPeeringConnectionModel) string { return m.VpcID.ValueString() },
		func(s *client.AWSEnvSpecFragment_PeeringConnections) string { return s.VpcID },
	)
	data.Id = data.Name
	data.Zones, diags = common.ListToModel(apiResp.CreateAWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.CreateAWSEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.CreateAWSEnv.SpecRevision)
	data.ResourcePrefix = types.StringValue(apiResp.CreateAWSEnv.Spec.ResourcePrefix)

	tflog.Trace(ctx, "created resource", map[string]interface{}{"name": envName})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *AWSEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AWSEnvResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "getting environment", map[string]interface{}{"name": envName})
	apiResp, err := r.Client.GetAWSEnv(ctx, envName)

	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "removing resource from state", map[string]interface{}{"name": envName})
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env %s, got error: %s", envName, client.FormatError(err, envName)))
		}
		return
	}

	// Reorder node groups, zones and tags to respect order in the user's configuration
	apiResp.AWSEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.AWSEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AWSEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.AWSEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.AWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.AWSEnv.Spec.Tags = common.ReorderByKey(data.Tags, apiResp.AWSEnv.Spec.Tags,
		func(m common.KeyValueModel) string { return m.Key.ValueString() },
		func(s *client.AWSEnvSpecFragment_Tags) string { return s.Key },
	)
	apiResp.AWSEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.AWSEnv.Spec.PeeringConnections,
		func(m AWSEnvPeeringConnectionModel) string { return m.VpcID.ValueString() },
		func(s *client.AWSEnvSpecFragment_PeeringConnections) string { return s.VpcID },
	)
	diags = data.toModel(*apiResp.AWSEnv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *AWSEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AWSEnvResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "updating resource", map[string]interface{}{"name": envName})

	_, sdkEnv, diags := data.toSDK(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiResp, err := r.Client.UpdateAWSEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	// Reorder node groups, zones and tags to respect order in the user's configuration
	apiResp.UpdateAWSEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.UpdateAWSEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AWSEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.UpdateAWSEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.UpdateAWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.UpdateAWSEnv.Spec.Tags = common.ReorderByKey(data.Tags, apiResp.UpdateAWSEnv.Spec.Tags,
		func(m common.KeyValueModel) string { return m.Key.ValueString() },
		func(s *client.AWSEnvSpecFragment_Tags) string { return s.Key },
	)
	apiResp.UpdateAWSEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.UpdateAWSEnv.Spec.PeeringConnections,
		func(m AWSEnvPeeringConnectionModel) string { return m.VpcID.ValueString() },
		func(s *client.AWSEnvSpecFragment_PeeringConnections) string { return s.VpcID },
	)
	data.Zones, diags = common.ListToModel(apiResp.UpdateAWSEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.UpdateAWSEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.UpdateAWSEnv.SpecRevision)
	data.ResourcePrefix = types.StringValue(apiResp.UpdateAWSEnv.Spec.ResourcePrefix)

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": envName})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *AWSEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AWSEnvResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	if !data.ForceDestroy.ValueBool() {
		resp.Diagnostics.AddError("Env Locked", fmt.Sprintf("env %s is protected for deletion, set `force_destroy` property to `true` and run `terraform apply` to unlock it", envName))
		return
	}

	envStatus, err := r.Client.GetAWSEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status  %s, got error: %s", envName, err))
		}
		return
	}

	if len(envStatus.AWSEnv.Status.Errors) > 0 {
		for _, err := range envStatus.AWSEnv.Status.Errors {
			if (err.Code == "DISCONNECTED" || err.Code == "K8S_DISCONNECTED") && !data.SkipDeprovisionOnDestroy.ValueBool() && !data.AllowDeleteWhileDisconnected.ValueBool() {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start env %s, environment is DISCONNECTED.\nCheck environment's `cloudconnect` or use `allow_delete_while_disconnected=true` to continue with the delete operation.", envName))
				return
			}
		}
	}

	apiResp, err := r.Client.DeleteAWSEnv(ctx, client.DeleteAWSEnvInput{
		Name:                 envName,
		Force:                data.SkipDeprovisionOnDestroy.ValueBoolPointer(),
		ForceDestroyClusters: data.ForceDestroyClusters.ValueBoolPointer(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", common.FormatDeleteError(envName, err))
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, common.DeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	common.WaitForDeletion(ctx, resp, envName, apiResp.DeleteAWSEnv.PendingMfa,
		func(ctx context.Context, name string) (bool, error) {
			status, err := r.Client.GetAWSEnvStatus(ctx, name)
			if err != nil {
				return false, err
			}
			return status.AWSEnv.Status.PendingDelete, nil
		},
		deleteTimeout,
		common.MFATimeout,
	)
}
