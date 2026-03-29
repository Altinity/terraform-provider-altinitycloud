package env

import (
	"context"
	"fmt"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &K8SEnvResource{}
var _ resource.ResourceWithImportState = &K8SEnvResource{}

func NewK8SEnvResource() resource.Resource {
	return &K8SEnvResource{}
}

type K8SEnvResource struct {
	common.EnvResourceBase
}

func (r *K8SEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_k8s"
}

func (r *K8SEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *K8SEnvResourceModel

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
	apiResp, err := r.Client.CreateK8SEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", name, client.FormatError(err, name)))
		return
	}

	// Reorder node groups  and zones to respect order in the user's configuration
	apiResp.CreateK8SEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.CreateK8SEnv.Spec.NodeGroups)
	data.Id = data.Name
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.CreateK8SEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.CreateK8SEnv.SpecRevision)
	diags = data.toModel(data.Name.ValueString(), apiResp.CreateK8SEnv.SpecRevision, *apiResp.CreateK8SEnv.Spec)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created resource", map[string]interface{}{"name": name})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *K8SEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *K8SEnvResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "getting environment", map[string]interface{}{"name": envName})
	apiResp, err := r.Client.GetK8SEnv(ctx, envName)

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
	apiResp.K8sEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.K8sEnv.Spec.NodeGroups)
	diags = data.toModel(apiResp.K8sEnv.Name, apiResp.K8sEnv.SpecRevision, *apiResp.K8sEnv.Spec)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *K8SEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *K8SEnvResourceModel

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
	apiResp, err := r.Client.UpdateK8SEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", name, client.FormatError(err, name)))
		return
	}

	apiResp.UpdateK8SEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.UpdateK8SEnv.Spec.NodeGroups)
	data.NodeGroups, diags = nodeGroupsToModel(apiResp.UpdateK8SEnv.Spec.NodeGroups)
	resp.Diagnostics.Append(diags...)
	data.SpecRevision = types.Int64Value(apiResp.UpdateK8SEnv.SpecRevision)

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": name})
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *K8SEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *K8SEnvResourceModel

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

	envStatus, err := r.Client.GetK8SEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status  %s, got error: %s", envName, err))
		}
		return
	}

	if len(envStatus.K8sEnv.Status.Errors) > 0 {
		for _, err := range envStatus.K8sEnv.Status.Errors {
			if err.Code == "DISCONNECTED" && !data.SkipDeprovisionOnDestroy.ValueBool() && !data.AllowDeleteWhileDisconnected.ValueBool() {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start env %s, environment is DISCONNECTED.\nCheck environment's `cloudconnect` or use `allow_delete_while_disconnected=true` to continue with the delete operation.", envName))
				return
			}
		}
	}

	apiResp, err := r.Client.DeleteK8SEnv(ctx, client.DeleteK8SEnvInput{
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

	common.WaitForDeletion(ctx, resp, envName, apiResp.DeleteK8SEnv.PendingMfa,
		func(ctx context.Context, name string) (bool, error) {
			status, err := r.Client.GetK8SEnvStatus(ctx, name)
			if err != nil {
				return false, err
			}
			return status.K8sEnv.Status.PendingDelete, nil
		},
		deleteTimeout,
		common.MFATimeout,
	)
}
