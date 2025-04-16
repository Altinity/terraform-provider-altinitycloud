package env

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

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

	sdkEnv, _ := data.toSDK()

	apiResp, err := r.Client.CreateHCloudEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", name, err))
		return
	}

	// Reorder node groups  and locations to respect order in the user's configuration
	apiResp.CreateHCloudEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.CreateHCloudEnv.Spec.NodeGroups)
	apiResp.CreateHCloudEnv.Spec.Locations = common.ReorderList(data.Locations, apiResp.CreateHCloudEnv.Spec.Locations)
	data.Id = data.Name
	data.Locations = common.ListToModel(apiResp.CreateHCloudEnv.Spec.Locations)
	data.NodeGroups = nodeGroupsToModel(apiResp.CreateHCloudEnv.Spec.NodeGroups)
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
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env %s, got error: %s", envName, err))
		}
		return
	}

	// Reorder node groups  and locations to respect order in the user's configuration
	apiResp.HcloudEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.HcloudEnv.Spec.NodeGroups)
	apiResp.HcloudEnv.Spec.Locations = common.ReorderList(data.Locations, apiResp.HcloudEnv.Spec.Locations)
	data.toModel(*apiResp.HcloudEnv)
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

	_, sdkEnv := data.toSDK()
	apiResp, err := r.Client.UpdateHCloudEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", name, err))
		return
	}

	// Reorder node groups  and locations to respect order in the user's configuration
	apiResp.UpdateHCloudEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.UpdateHCloudEnv.Spec.NodeGroups)
	apiResp.UpdateHCloudEnv.Spec.Locations = common.ReorderList(data.Locations, apiResp.UpdateHCloudEnv.Spec.Locations)
	data.Locations = common.ListToModel(apiResp.UpdateHCloudEnv.Spec.Locations)
	data.NodeGroups = nodeGroupsToModel(apiResp.UpdateHCloudEnv.Spec.NodeGroups)
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
	if !data.ForceDestroy.ValueBool() {
		resp.Diagnostics.AddError("Env Locked", fmt.Sprintf("env %s is protected for deletion, set `force_destroy` property to `true` and run `terraform apply` to unlock it", envName))
		return
	}

	envStatus, err := r.Client.GetHCloudEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status  %s, got error: %s", envName, err))
		}
		return
	}

	if len(envStatus.HcloudEnv.Status.Errors) > 0 {
		for _, err := range envStatus.HcloudEnv.Status.Errors {
			if err.Code == "DISCONNECTED" && !data.SkipDeprovisionOnDestroy.ValueBool() && !data.AllowDeleteWhileDisconnected.ValueBool() {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start env %s, environment is DISCONNECTED.\nCheck environment's `cloudconnect` or use `allow_delete_while_disconnected=true` to continue with the delete operation.", envName))
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
		errMessage := fmt.Sprintf("Unable to delete env %s, got error: %s", envName, err)
		activeClustes, _ := client.IsActiceClustersError(err)
		if activeClustes {
			errMessage = fmt.Sprintf("Unable to delete env %s, it has active ClickHouse/Zookeeper clusters (use force_destroy_clusters=true to force delete them)", envName)
		}

		resp.Diagnostics.AddError("Client Error", errMessage)
		return
	}

	// Polling to wait for deletion to complete
	pendingMfa := apiResp.DeleteHCloudEnv.PendingMfa
	mfaTimeout := time.After(common.MFA_TIMEOUT)
	deleteTimeout := time.After(common.DELETE_TIMEOUT)
	ticker := time.NewTicker(common.DELETE_POLL_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Context Cancelled", "The context was cancelled, stopping env deletion.")
			return
		case <-mfaTimeout:
			if pendingMfa {
				resp.Diagnostics.AddError("MFA Timeout", "Timeout reached while waiting for MFA to be confirmed.\nPlease check your MFA device, confirm deletion and run `terraform destroy` again.")
				return
			}
		case <-deleteTimeout:
			resp.Diagnostics.AddError("Timeout", "Timeout reached while waiting for env to be deleted.")
			return
		case <-ticker.C:
			tflog.Trace(ctx, "checking if env was deleted", map[string]interface{}{"name": envName})
			envStatus, err := r.Client.GetHCloudEnvStatus(ctx, envName)
			pendingMfa = !envStatus.HcloudEnv.Status.PendingDelete

			if err != nil {
				notFound, err := client.IsNotFoundError(err)
				if notFound {
					tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
				} else {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env %s, got error: %s", envName, err))
				}
				return
			}
		}
	}
}
