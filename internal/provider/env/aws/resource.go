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

	sdkEnv, _ := data.toSDK()
	apiResp, err := r.Client.CreateAWSEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", envName, err))
		return
	}

	// Reorder node groups  and zones to respect order in the user's configuration
	apiResp.CreateAWSEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.CreateAWSEnv.Spec.NodeGroups)
	apiResp.CreateAWSEnv.Spec.Zones = common.ReorderList(data.Zones, apiResp.CreateAWSEnv.Spec.Zones)
	data.Id = data.Name
	data.Zones = common.ListToModel(apiResp.CreateAWSEnv.Spec.Zones)
	data.NodeGroups = nodeGroupsToModel(apiResp.CreateAWSEnv.Spec.NodeGroups)
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
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env %s, got error: %s", envName, err))
		}
		return
	}

	// Reorder node groups  and zones to respect order in the user's configuration
	apiResp.AWSEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.AWSEnv.Spec.NodeGroups)
	apiResp.AWSEnv.Spec.Zones = common.ReorderList(data.Zones, apiResp.AWSEnv.Spec.Zones)
	data.toModel(*apiResp.AWSEnv)
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

	_, sdkEnv := data.toSDK()
	apiResp, err := r.Client.UpdateAWSEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", envName, err))
		return
	}

	// Reorder node groups  and zones to respect order in the user's configuration
	apiResp.UpdateAWSEnv.Spec.NodeGroups = reorderNodeGroups(data.NodeGroups, apiResp.UpdateAWSEnv.Spec.NodeGroups)
	apiResp.UpdateAWSEnv.Spec.Zones = common.ReorderList(data.Zones, apiResp.UpdateAWSEnv.Spec.Zones)
	data.Zones = common.ListToModel(apiResp.UpdateAWSEnv.Spec.Zones)
	data.NodeGroups = nodeGroupsToModel(apiResp.UpdateAWSEnv.Spec.NodeGroups)
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
			if err.Code == "DISCONNECTED" && !data.SkipDeprovisionOnDestroy.ValueBool() && !data.AllowDeleteWhileDisconnected.ValueBool() {
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
		errMessage := fmt.Sprintf("Unable to delete env %s, got error: %s", envName, err)
		activeClustes, _ := client.IsActiceClustersError(err)
		if activeClustes {
			errMessage = fmt.Sprintf("Unable to delete env %s, it has active ClickHouse/Zookeeper clusters (use force_destroy_clusters=true to force delete them)", envName)
		}

		resp.Diagnostics.AddError("Client Error", errMessage)
		return
	}

	// Polling to wait for deletion to complete
	pendingMfa := apiResp.DeleteAWSEnv.PendingMfa
	mfaTimeout := time.After(common.MFA_TIMEOUT)
	deleteTimeout := time.After(common.DELETE_TIMEOUT)
	ticker := time.NewTicker(common.DELETE_POLL_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Context Cancelled", "The context was cancelled, stopping env deletion.")
			return
		case <-deleteTimeout:
			resp.Diagnostics.AddError("Timeout", "Timeout reached while waiting for env to be deleted.")
			return
		case <-mfaTimeout:
			if pendingMfa {
				resp.Diagnostics.AddError("MFA Timeout", "Timeout reached while waiting for MFA to be confirmed.\nPlease check your MFA device, confirm deletion and run `terraform destroy` again.")
				return
			}
		case <-ticker.C:
			tflog.Trace(ctx, "checking if env was deleted", map[string]interface{}{"name": envName})
			envStatus, err := r.Client.GetAWSEnvStatus(ctx, envName)
			pendingMfa = !envStatus.AWSEnv.Status.PendingDelete

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
