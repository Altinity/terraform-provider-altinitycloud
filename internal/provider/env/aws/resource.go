package env

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/auth"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &AWSEnvResource{}
var _ resource.ResourceWithImportState = &AWSEnvResource{}

func NewAWSEnvResource() resource.Resource {
	return &AWSEnvResource{}
}

type AWSEnvResource struct {
	client *client.Client
	auth   *auth.Auth
}

func (r *AWSEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_aws"
}

func (r *AWSEnvResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*sdk.AltinityCloudSDK)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = sdk.Client
	r.auth = sdk.Auth
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
	apiResp, err := r.client.CreateAWSEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create env %s, got error: %s", envName, err))
		return
	}

	data.Id = data.Name
	data.Zones = common.ListToModel(apiResp.CreateAWSEnv.Spec.Zones)
	data.NodeGroups = nodeGroupsToModel(apiResp.CreateAWSEnv.Spec.NodeGroups)
	data.SpecRevision = types.Int64Value(apiResp.CreateAWSEnv.SpecRevision)

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
	apiResp, err := r.client.GetAWSEnv(ctx, envName)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env %s, got error: %s", envName, err))
		return
	}

	if apiResp.AwsEnv == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	data.toModel(*apiResp.AwsEnv)
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
	apiResp, err := r.client.UpdateAWSEnv(ctx, sdkEnv)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update env %s, got error: %s", envName, err))
		return
	}

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": envName})

	data.Zones = common.ListToModel(apiResp.UpdateAWSEnv.Spec.Zones)
	data.NodeGroups = nodeGroupsToModel(apiResp.UpdateAWSEnv.Spec.NodeGroups)
	data.SpecRevision = types.Int64Value(apiResp.UpdateAWSEnv.SpecRevision)

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

	_, err := r.client.DeleteAWSEnv(ctx, client.DeleteAWSEnvInput{
		Name:                 envName,
		Force:                data.SkipDeprovisionOnDestroy.ValueBoolPointer(),
		ForceDestroyClusters: data.ForceDestroyClusters.ValueBoolPointer(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete env %s, got error: %s", envName, err))
		return
	}

	_, err = r.client.GetAWSEnv(ctx, envName)
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
	timeout := time.After(time.Duration(30) * time.Minute)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Context Cancelled", "The context was cancelled, stopping env deletion.")
			return
		case <-timeout:
			resp.Diagnostics.AddError("Timeout", "Timeout reached while waiting for env to be deleted.")
			return
		case <-ticker.C:
			tflog.Trace(ctx, "checking if env was deleted", map[string]interface{}{"name": envName})
			_, err := r.client.GetAWSEnv(ctx, envName)

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

func (r *AWSEnvResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r AWSEnvResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		var skipDeprovision types.Bool
		req.State.GetAttribute(ctx, path.Root("skip_deprovision_on_destroy"), &skipDeprovision)

		if skipDeprovision.ValueBool() {
			resp.Diagnostics.AddAttributeWarning(path.Root("skip_deprovision_on_destroy"), "Skip Deprovision on Destroy", "This resource is using the 'skip_deprovision_on_destroy'.\nUse this with precaution as it will delete the environment without deleting any of your cloud resources.")
		}
	}
}
