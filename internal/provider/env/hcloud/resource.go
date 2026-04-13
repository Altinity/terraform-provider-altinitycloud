package env

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	support "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &HCloudEnvResource{}
var _ resource.ResourceWithImportState = &HCloudEnvResource{}

func NewHCloudEnvResource() resource.Resource {
	return &HCloudEnvResource{}
}

type HCloudEnvResource struct {
	common.EnvResourceBase
}

func (r *HCloudEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_hcloud"
}

func (r *HCloudEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *HCloudEnvResourceModel

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

	apiResp, err := r.Client.CreateHCloudEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", support.FormatClientError(fmt.Sprintf("Unable to create env %s, got error: %s", name, client.FormatError(err, name))))
		return
	}

	// Reorder node groups  and locations to respect order in the user's configuration
	apiResp.CreateHCloudEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.CreateHCloudEnv.Spec.NodeGroups,
		func(m NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.HCloudEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.CreateHCloudEnv.Spec.Locations, diags = common.ReorderList(ctx, data.Locations, apiResp.CreateHCloudEnv.Spec.Locations)
	resp.Diagnostics.Append(diags...)
	data.Id = data.Name
	data.Locations, diags = common.ListToModel(apiResp.CreateHCloudEnv.Spec.Locations)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.CreateHCloudEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.CreateHCloudEnv.SpecRevision)

	tflog.Trace(ctx, "created resource", map[string]interface{}{"name": name})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *HCloudEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *HCloudEnvResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "getting environment", map[string]interface{}{"name": envName})
	apiResp, err := r.Client.GetHCloudEnv(ctx, envName)

	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "removing resource from state", map[string]interface{}{"name": envName})
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", support.FormatClientError(fmt.Sprintf("Unable to read env %s, got error: %s", envName, client.FormatError(err, envName))))
		}
		return
	}

	// Reorder node groups  and locations to respect order in the user's configuration
	apiResp.HcloudEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.HcloudEnv.Spec.NodeGroups,
		func(m NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.HCloudEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.HcloudEnv.Spec.Locations, diags = common.ReorderList(ctx, data.Locations, apiResp.HcloudEnv.Spec.Locations)
	resp.Diagnostics.Append(diags...)
	diags = data.toModel(*apiResp.HcloudEnv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *HCloudEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *HCloudEnvResourceModel

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
	apiResp, err := r.Client.UpdateHCloudEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", support.FormatClientError(fmt.Sprintf("Unable to update env %s, got error: %s", name, client.FormatError(err, name))))
		return
	}

	// Reorder node groups  and locations to respect order in the user's configuration
	apiResp.UpdateHCloudEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.UpdateHCloudEnv.Spec.NodeGroups,
		func(m NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.HCloudEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.UpdateHCloudEnv.Spec.Locations, diags = common.ReorderList(ctx, data.Locations, apiResp.UpdateHCloudEnv.Spec.Locations)
	resp.Diagnostics.Append(diags...)
	data.Locations, diags = common.ListToModel(apiResp.UpdateHCloudEnv.Spec.Locations)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.UpdateHCloudEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.UpdateHCloudEnv.SpecRevision)

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": name})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *HCloudEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *HCloudEnvResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	resp.Diagnostics.Append(common.ValidateForceDestroy(envName, data.ForceDestroy.ValueBool())...)
	if resp.Diagnostics.HasError() {
		return
	}

	envStatus, err := r.Client.GetHCloudEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			resp.Diagnostics.AddError("Client Error", support.FormatClientError(fmt.Sprintf("Unable to read env status %s, got error: %s", envName, err)))
		}
		return
	}

	if len(envStatus.HcloudEnv.Status.Errors) > 0 {
		for _, err := range envStatus.HcloudEnv.Status.Errors {
			resp.Diagnostics.Append(common.ValidateDisconnected(
				envName,
				string(err.Code),
				envStatus.HcloudEnv.Status.AppliedSpecRevision,
				data.SkipDeprovisionOnDestroy.ValueBool(),
				data.AllowDeleteWhileDisconnected.ValueBool(),
			)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	apiResp, err := r.Client.DeleteHCloudEnv(ctx, client.DeleteHCloudEnvInput{
		Name:                 envName,
		Force:                data.SkipDeprovisionOnDestroy.ValueBoolPointer(),
		ForceDestroyClusters: data.ForceDestroyClusters.ValueBoolPointer(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", support.FormatClientError(common.FormatDeleteError(envName, err)))
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, common.DeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	common.WaitForDeletion(ctx, resp, envName, apiResp.DeleteHCloudEnv.PendingMfa,
		func(ctx context.Context, name string) (bool, error) {
			status, err := r.Client.GetHCloudEnvStatus(ctx, name)
			if err != nil {
				return false, err
			}
			return status.HcloudEnv.Status.PendingDelete, nil
		},
		deleteTimeout,
		common.MFATimeout,
	)
}
