package hosted_env

import (
	"context"
	"fmt"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &HostedAWSEnvResource{}
var _ resource.ResourceWithImportState = &HostedAWSEnvResource{}
var _ resource.ResourceWithValidateConfig = &HostedAWSEnvResource{}

func NewHostedAWSEnvResource() resource.Resource {
	return &HostedAWSEnvResource{}
}

type HostedAWSEnvResource struct {
	common.EnvResourceBase
}

func (r *HostedAWSEnvResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosted_env_aws"
}

func (r *HostedAWSEnvResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	resp.Diagnostics.Append(common.ValidateNodeGroupNames(ctx, req.Config)...)

	// Read only datadog: a full Config.Get panics on unknown nested struct-pointer attrs.
	var datadogObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("datadog"), &datadogObj)...)
	if resp.Diagnostics.HasError() || datadogObj.IsNull() || datadogObj.IsUnknown() {
		return
	}

	var datadog *common.DatadogModel
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("datadog"), &datadog)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(common.ValidateDatadog(datadog)...)
}

func (r *HostedAWSEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *HostedAWSEnvResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
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

	apiResp, err := r.Client.CreateHostedAWSEnv(ctx, sdkEnv)
	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to create env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	resp.Diagnostics.Append(data.applySpec(ctx, envName, apiResp.CreateHostedAWSEnv.Spec, apiResp.CreateHostedAWSEnv.SpecRevision)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	tflog.Trace(ctx, "created resource", map[string]interface{}{"name": envName})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostedAWSEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *HostedAWSEnvResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	tflog.Trace(ctx, "getting environment", map[string]interface{}{"name": envName})

	apiResp, err := r.Client.GetHostedAWSEnv(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "removing resource from state", map[string]interface{}{"name": envName})
			resp.State.RemoveResource(ctx)
		} else {
			clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env %s, got error: %s", envName, client.FormatError(err, envName)))
		}
		return
	}

	if apiResp.HostedAWSEnv == nil {
		tflog.Trace(ctx, "removing resource from state", map[string]interface{}{"name": envName})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(data.applySpec(ctx, apiResp.HostedAWSEnv.Name, apiResp.HostedAWSEnv.Spec, apiResp.HostedAWSEnv.SpecRevision)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostedAWSEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *HostedAWSEnvResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
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

	apiResp, err := r.Client.UpdateHostedAWSEnv(ctx, sdkEnv)
	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to update env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	resp.Diagnostics.Append(data.applySpec(ctx, envName, apiResp.UpdateHostedAWSEnv.Spec, apiResp.UpdateHostedAWSEnv.SpecRevision)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated resource", map[string]interface{}{"name": envName})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostedAWSEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *HostedAWSEnvResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	resp.Diagnostics.Append(common.ValidateForceDestroy(envName, data.ForceDestroy.ValueBool())...)
	if resp.Diagnostics.HasError() {
		return
	}

	envStatus, err := r.Client.GetHostedAWSEnvStatus(ctx, envName)
	if err != nil {
		notFound, _ := client.IsNotFoundError(err)
		if notFound {
			tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		} else {
			clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env status %s, got error: %s", envName, err))
		}
		return
	}

	// A successful query still returns a null env when it is already gone.
	if envStatus.HostedAWSEnv == nil {
		tflog.Trace(ctx, "deleted resource", map[string]interface{}{"name": envName})
		return
	}

	for _, statusErr := range envStatus.HostedAWSEnv.Status.Errors {
		resp.Diagnostics.Append(common.ValidateDisconnected(
			envName,
			string(statusErr.Code),
			envStatus.HostedAWSEnv.Status.AppliedSpecRevision,
			data.SkipDeprovisionOnDestroy.ValueBool(),
			data.AllowDeleteWhileDisconnected.ValueBool(),
		)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	apiResp, err := r.Client.DeleteHostedAWSEnv(ctx, client.DeleteHostedAWSEnvInput{
		Name:                 envName,
		Force:                data.SkipDeprovisionOnDestroy.ValueBoolPointer(),
		ForceDestroyClusters: data.ForceDestroyClusters.ValueBoolPointer(),
	})
	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, common.FormatDeleteError(envName, err))
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, common.DeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	common.WaitForDeletion(ctx, resp, envName, apiResp.DeleteHostedAWSEnv.PendingMfa,
		func(ctx context.Context, name string) (bool, error) {
			status, err := r.Client.GetHostedAWSEnvStatus(ctx, name)
			if err != nil {
				return false, err
			}
			if status.HostedAWSEnv == nil {
				return false, common.ErrEnvNotFound
			}
			return status.HostedAWSEnv.Status.PendingDelete, nil
		},
		deleteTimeout,
		common.MFATimeout,
	)
}
