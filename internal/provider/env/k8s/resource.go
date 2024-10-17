package env

import (
	"context"
	"fmt"
	"time"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/auth"
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
	client *client.Client
	auth   *auth.Auth
}

func (r *K8SEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_k8s"
}

func (r *K8SEnvResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = sdk.Client
	r.auth = sdk.Auth
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

	sdkEnv, _ := data.toSDK()
	apiResp, err := r.client.CreateK8SEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", name, err))
		return
	}

	data.Id = data.Name
	data.NodeGroups = nodeGroupsToModel(apiResp.CreateK8SEnv.Spec.NodeGroups)
	data.SpecRevision = types.Int64Value(apiResp.CreateK8SEnv.SpecRevision)

	data.toModel(data.Name.ValueString(), *apiResp.CreateK8SEnv.Spec)
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
	apiResp, err := r.client.GetK8SEnv(ctx, envName)

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

	data.toModel(apiResp.K8sEnv.Name, *apiResp.K8sEnv.Spec)
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

	_, sdkEnv := data.toSDK()
	apiResp, err := r.client.UpdateK8SEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", name, err))
		return
	}

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": name})
	data.NodeGroups = nodeGroupsToModel(apiResp.UpdateK8SEnv.Spec.NodeGroups)
	data.SpecRevision = types.Int64Value(apiResp.UpdateK8SEnv.SpecRevision)

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

	apiResp, err := r.client.DeleteK8SEnv(ctx, client.DeleteK8SEnvInput{
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

	pendingMfa := apiResp.DeleteK8SEnv.PendingMfa
	_, err = r.client.GetK8SEnvStatus(ctx, envName)
	if err != nil {
		notFound, err := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env %s, got error: %s", envName, err))
		}
		return
	}

	// Polling to wait for deletion to complete
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
			envStatus, err := r.client.GetK8SEnvStatus(ctx, envName)
			pendingMfa = !envStatus.K8sEnv.Status.PendingDelete

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

func (r *K8SEnvResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	common.Import(ctx, req, resp)
}

func (r K8SEnvResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	common.ModifyPlan(ctx, req, resp)
}
