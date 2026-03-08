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

var _ resource.Resource = &AzureEnvResource{}
var _ resource.ResourceWithImportState = &AzureEnvResource{}

func NewAzureEnvResource() resource.Resource {
	return &AzureEnvResource{}
}

type AzureEnvResource struct {
	common.EnvResourceBase
}

func (r *AzureEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_azure"
}

func (r *AzureEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AzureEnvResourceModel

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

	apiResp, err := r.Client.CreateAzureEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", name, client.FormatError(err, name)))
		return
	}

	// Reorder node groups  and zones to respect order in the user's configuration
	apiResp.CreateAzureEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.CreateAzureEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AzureEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.CreateAzureEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.CreateAzureEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.Id = data.Name
	data.Zones, diags = common.ListToModel(apiResp.CreateAzureEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.CreateAzureEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.CreateAzureEnv.SpecRevision)

	tflog.Trace(ctx, "created resource", map[string]interface{}{"name": name})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *AzureEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AzureEnvResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "getting environment", map[string]interface{}{"name": envName})
	apiResp, err := r.Client.GetAzureEnv(ctx, envName)

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

	// Reorder node groups  and zones to respect order in the user's configuration
	apiResp.AzureEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.AzureEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AzureEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.AzureEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.AzureEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	diags = data.toModel(*apiResp.AzureEnv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *AzureEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AzureEnvResourceModel

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
	apiResp, err := r.Client.UpdateAzureEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", name, client.FormatError(err, name)))
		return
	}

	// Reorder node groups  and zones to respect order in the user's configuration
	apiResp.UpdateAzureEnv.Spec.NodeGroups = common.ReorderByKey(data.NodeGroups, apiResp.UpdateAzureEnv.Spec.NodeGroups,
		func(m common.NodeGroupsModel) string { return m.NodeType.ValueString() },
		func(s *client.AzureEnvSpecFragment_NodeGroups) string { return s.NodeType },
	)
	apiResp.UpdateAzureEnv.Spec.Zones, diags = common.ReorderList(ctx, data.Zones, apiResp.UpdateAzureEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.Zones, diags = common.ListToModel(apiResp.UpdateAzureEnv.Spec.Zones)
	resp.Diagnostics.Append(diags...)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.UpdateAzureEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.UpdateAzureEnv.SpecRevision)

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": name})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *AzureEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AzureEnvResourceModel

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

	envStatus, err := r.Client.GetAzureEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status  %s, got error: %s", envName, err))
		}
		return
	}

	if len(envStatus.AzureEnv.Status.Errors) > 0 {
		for _, err := range envStatus.AzureEnv.Status.Errors {
			if err.Code == "DISCONNECTED" && !data.SkipDeprovisionOnDestroy.ValueBool() && !data.AllowDeleteWhileDisconnected.ValueBool() {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start env %s, environment is DISCONNECTED.\nCheck environment's `cloudconnect` or use `allow_delete_while_disconnected=true` to continue with the delete operation.", envName))
				return
			}
		}
	}

	apiResp, err := r.Client.DeleteAzureEnv(ctx, client.DeleteAzureEnvInput{
		Name:                 envName,
		Force:                data.SkipDeprovisionOnDestroy.ValueBoolPointer(),
		ForceDestroyClusters: data.ForceDestroyClusters.ValueBoolPointer(),
	})

	if err != nil {
		errMessage := fmt.Sprintf("Unable to delete env %s, got error: %s", envName, err)
		activeClustes, _ := client.IsActiveClustersError(err)
		if activeClustes {
			errMessage = fmt.Sprintf("Unable to delete env %s, it has active ClickHouse/Zookeeper clusters (use force_destroy_clusters=true to force delete them)", envName)
		}

		resp.Diagnostics.AddError("Client Error", errMessage)
		return
	}

	common.WaitForDeletion(ctx, resp, envName, apiResp.DeleteAzureEnv.PendingMfa,
		func(ctx context.Context, name string) (bool, error) {
			status, err := r.Client.GetAzureEnvStatus(ctx, name)
			if err != nil {
				return false, err
			}
			return status.AzureEnv.Status.PendingDelete, nil
		},
	)
}
