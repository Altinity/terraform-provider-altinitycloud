package env_status

import (
	"context"
	"fmt"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &AWSEnvStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &AWSEnvStatusDataSource{}

func NewAWSEnvStatusDataSource() datasource.DataSource {
	return &AWSEnvStatusDataSource{}
}

type AWSEnvStatusDataSource struct {
	common.EnvStatusDataSourceBase
}

func (d *AWSEnvStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_aws_status"
	d.TypeName = resp.TypeName
}

func (d *AWSEnvStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading aws env status data source")

	var data AWSEnvStatusModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.Client.GetAWSEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.AWSEnv == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	waitForAppliedSpecRevision := data.WaitForAppliedSpecRevision.ValueInt64()
	if waitForAppliedSpecRevision == 0 || apiResp.AWSEnv.Status.AppliedSpecRevision >= waitForAppliedSpecRevision {
		tflog.Trace(ctx, "env status matchs spec", map[string]interface{}{"name": envName})
		data.toModel(*apiResp.AWSEnv)
		data.Id = data.Name

		diags = resp.State.Set(ctx, &data)
		resp.Diagnostics.Append(diags...)

		return
	}

	poll := func(ctx context.Context, name string) (*common.PollResult, error) {
		resp, err := d.Client.GetAWSEnvStatus(ctx, name)
		if err != nil {
			return nil, err
		}
		if resp.AWSEnv == nil {
			return &common.PollResult{Found: false}, nil
		}
		var errors []common.EnvError
		for _, e := range resp.AWSEnv.Status.Errors {
			errors = append(errors, common.EnvError{Code: string(e.Code), Message: e.Message})
		}
		return &common.PollResult{
			AppliedSpecRevision: resp.AWSEnv.Status.AppliedSpecRevision,
			Errors:              errors,
			Found:               true,
		}, nil
	}

	readTimeout, diags := data.Timeouts.Read(ctx, common.MATCH_SPEC_TIMEOUT)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !common.WaitForSpecRevision(ctx, envName, d.TypeName, waitForAppliedSpecRevision, data.Verbose.ValueBool(), poll, &resp.Diagnostics, readTimeout) {
		return
	}

	// Re-fetch to populate the model with latest data
	apiResp, err = d.Client.GetAWSEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	data.toModel(*apiResp.AWSEnv)
	data.Id = data.Name
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
