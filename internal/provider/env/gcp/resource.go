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

	sdkEnv, _ := data.toSDK()

	apiResp, err := r.Client.CreateGCPEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", name, err))
		return
	}

	tflog.Trace(ctx, "created resource", map[string]interface{}{"name": name})
	data.Id = data.Name
	data.Zones = common.ListToModel(apiResp.CreateGCPEnv.Spec.Zones)
	data.NodeGroups = nodeGroupsToModel(apiResp.CreateGCPEnv.Spec.NodeGroups)
	data.SpecRevision = types.Int64Value(apiResp.CreateGCPEnv.SpecRevision)

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
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env %s, got error: %s", envName, err))
		}
		return
	}

	data.toModel(*apiResp.GcpEnv)
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

	_, sdkEnv := data.toSDK()
	apiResp, err := r.Client.UpdateGCPEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", name, err))
		return
	}

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": name})
	data.Zones = common.ListToModel(apiResp.UpdateGCPEnv.Spec.Zones)
	data.NodeGroups = nodeGroupsToModel(apiResp.UpdateGCPEnv.Spec.NodeGroups)
	data.SpecRevision = types.Int64Value(apiResp.UpdateGCPEnv.SpecRevision)

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

	apiResp, err := r.Client.DeleteGCPEnv(ctx, client.DeleteGCPEnvInput{
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

	pendingMfa := apiResp.DeleteGCPEnv.PendingMfa
	_, err = r.Client.GetGCPEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
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
			envStatus, err := r.Client.GetGCPEnvStatus(ctx, envName)
			pendingMfa = !envStatus.GcpEnv.Status.PendingDelete

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
