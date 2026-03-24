package env_status

import (
	"context"
	"fmt"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &AzureEnvStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &AzureEnvStatusDataSource{}

func NewAzureEnvStatusDataSource() datasource.DataSource {
	return &AzureEnvStatusDataSource{}
}

type AzureEnvStatusDataSource struct {
	common.EnvStatusDataSourceBase
}

func (d *AzureEnvStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_azure_status"
}

func (d *AzureEnvStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading azure env status data source")

	var data AzureEnvStatusModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.Client.GetAzureEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.AzureEnv == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	waitForAppliedSpecRevision := data.WaitForAppliedSpecRevision.ValueInt64()
	if waitForAppliedSpecRevision == 0 || apiResp.AzureEnv.Status.AppliedSpecRevision >= waitForAppliedSpecRevision {
		tflog.Trace(ctx, "env status matchs spec", map[string]interface{}{"name": envName})
		data.toModel(*apiResp.AzureEnv)
		data.Id = data.Name

		diags = resp.State.Set(ctx, &data)
		resp.Diagnostics.Append(diags...)

		return
	}

	poll := func(ctx context.Context, name string) (*common.PollResult, error) {
		resp, err := d.Client.GetAzureEnvStatus(ctx, name)
		if err != nil {
			return nil, err
		}
		if resp.AzureEnv == nil {
			return &common.PollResult{Found: false}, nil
		}
		var errors []common.EnvError
		for _, e := range resp.AzureEnv.Status.Errors {
			errors = append(errors, common.EnvError{Code: string(e.Code), Message: e.Message})
		}
		return &common.PollResult{
			AppliedSpecRevision: resp.AzureEnv.Status.AppliedSpecRevision,
			Errors:              errors,
			Found:               true,
		}, nil
	}

	if !common.WaitForSpecRevision(ctx, envName, waitForAppliedSpecRevision, data.Verbose.ValueBool(), poll, &resp.Diagnostics) {
		return
	}

	apiResp, err = d.Client.GetAzureEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	data.toModel(*apiResp.AzureEnv)
	data.Id = data.Name
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
