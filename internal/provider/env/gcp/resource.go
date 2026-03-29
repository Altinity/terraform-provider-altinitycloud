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

var _ resource.Resource = &GCPEnvResource{}
var _ resource.ResourceWithImportState = &GCPEnvResource{}

func NewGCPEnvResource() resource.Resource {
	return &GCPEnvResource{}
}

type GCPEnvResource struct {
	common.EnvResourceBase
}

func (r *GCPEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_gcp"
}

func (r *GCPEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *GCPEnvResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tflog.Trace(ctx, "creating resource", map[string]interface{}{"name": name})

	sdkEnv, _, diags := data.toSDK(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.Client.CreateGCPEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", name, client.FormatError(err, name)))
		return
	}

	// Reorder node groups, zones and peering connections to respect order in the user's configuration
	apiResp.CreateGCPEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.CreateGCPEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.GCPEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.CreateGCPEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.CreateGCPEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.CreateGCPEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.CreateGCPEnv.Spec.PeeringConnections,
		func(m GCPEnvPeeringConnectionModel) string { return m.NetworkName.ValueString() },
		func(s *client.GCPEnvSpecFragment_PeeringConnections) string { return s.NetworkName },
	)
	data.Id = data.Name
	data.Zones, diags = common.ListToModel(apiResp.CreateGCPEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.CreateGCPEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.CreateGCPEnv.SpecRevision)

	tflog.Trace(ctx, "created resource", map[string]interface{}{"name": name})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *GCPEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *GCPEnvResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "getting environment", map[string]interface{}{"name": envName})
	apiResp, err := r.Client.GetGCPEnv(ctx, envName)

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

	// Reorder node groups, zones and peering connections to respect order in the user's configuration
	apiResp.GCPEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.GCPEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.GCPEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.GCPEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.GCPEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.GCPEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.GCPEnv.Spec.PeeringConnections,
		func(m GCPEnvPeeringConnectionModel) string { return m.NetworkName.ValueString() },
		func(s *client.GCPEnvSpecFragment_PeeringConnections) string { return s.NetworkName },
	)
	diags = data.toModel(*apiResp.GCPEnv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *GCPEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *GCPEnvResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tflog.Trace(ctx, "updating resource", map[string]interface{}{"name": name})

	_, sdkEnv, diags := data.toSDK(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiResp, err := r.Client.UpdateGCPEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", name, client.FormatError(err, name)))
		return
	}

	// Reorder node groups, zones and peering connections to respect order in the user's configuration
	apiResp.UpdateGCPEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.UpdateGCPEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.GCPEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.UpdateGCPEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.UpdateGCPEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	apiResp.UpdateGCPEnv.Spec.PeeringConnections = common.ReorderByKey(data.PeeringConnections, apiResp.UpdateGCPEnv.Spec.PeeringConnections,
		func(m GCPEnvPeeringConnectionModel) string { return m.NetworkName.ValueString() },
		func(s *client.GCPEnvSpecFragment_PeeringConnections) string { return s.NetworkName },
	)
	data.Zones, diags = common.ListToModel(apiResp.UpdateGCPEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.UpdateGCPEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.UpdateGCPEnv.SpecRevision)

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": name})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *GCPEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *GCPEnvResourceModel

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

	envStatus, err := r.Client.GetGCPEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status  %s, got error: %s", envName, err))
		}
		return
	}

	if len(envStatus.GCPEnv.Status.Errors) > 0 {
		for _, err := range envStatus.GCPEnv.Status.Errors {
			if err.Code == "DISCONNECTED" && !data.SkipDeprovisionOnDestroy.ValueBool() && !data.AllowDeleteWhileDisconnected.ValueBool() {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start env %s, environment is DISCONNECTED.\nCheck environment's `cloudconnect` or use `allow_delete_while_disconnected=true` to continue with the delete operation.", envName))
				return
			}
		}
	}

	apiResp, err := r.Client.DeleteGCPEnv(ctx, client.DeleteGCPEnvInput{
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

	common.WaitForDeletion(ctx, resp, envName, apiResp.DeleteGCPEnv.PendingMfa,
		func(ctx context.Context, name string) (bool, error) {
			status, err := r.Client.GetGCPEnvStatus(ctx, name)
			if err != nil {
				return false, err
			}
			return status.GCPEnv.Status.PendingDelete, nil
		},
		deleteTimeout,
		common.MFATimeout,
	)
}
