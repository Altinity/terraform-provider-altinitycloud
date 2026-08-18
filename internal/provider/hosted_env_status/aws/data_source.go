package hosted_env_status

import (
	"context"
	"fmt"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &HostedAWSEnvStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &HostedAWSEnvStatusDataSource{}

func NewHostedAWSEnvStatusDataSource() datasource.DataSource {
	return &HostedAWSEnvStatusDataSource{}
}

type HostedAWSEnvStatusDataSource struct {
	common.EnvStatusDataSourceBase
}

func (d *HostedAWSEnvStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosted_env_aws_status"
}

func (d *HostedAWSEnvStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading hosted aws env status data source")

	var data HostedAWSEnvStatusModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.Client.GetHostedAWSEnvStatus(ctx, envName)
	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.HostedAWSEnv == nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	waitForAppliedSpecRevision := data.WaitForAppliedSpecRevision.ValueInt64()
	if waitForAppliedSpecRevision == 0 || apiResp.HostedAWSEnv.Status.AppliedSpecRevision >= waitForAppliedSpecRevision {
		tflog.Trace(ctx, "env status matches spec", map[string]interface{}{"name": envName})
		data.toModel(*apiResp.HostedAWSEnv)
		data.Id = data.Name

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	poll := func(ctx context.Context, name string) (*common.PollResult, error) {
		statusResp, err := d.Client.GetHostedAWSEnvStatus(ctx, name)
		if err != nil {
			return nil, err
		}
		if statusResp.HostedAWSEnv == nil {
			return &common.PollResult{Found: false}, nil
		}
		var errors []common.EnvError
		for _, e := range statusResp.HostedAWSEnv.Status.Errors {
			errors = append(errors, common.EnvError{Code: string(e.Code), Message: e.Message})
		}
		return &common.PollResult{
			AppliedSpecRevision: statusResp.HostedAWSEnv.Status.AppliedSpecRevision,
			Errors:              errors,
			Found:               true,
		}, nil
	}

	readTimeout, diags := data.Timeouts.Read(ctx, common.MATCH_SPEC_TIMEOUT)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !common.WaitForSpecRevision(ctx, envName, waitForAppliedSpecRevision, data.Verbose.ValueBool(), poll, &resp.Diagnostics, readTimeout) {
		return
	}

	// Re-fetch to populate the model with latest data
	apiResp, err = d.Client.GetHostedAWSEnvStatus(ctx, envName)
	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.HostedAWSEnv == nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	data.toModel(*apiResp.HostedAWSEnv)
	data.Id = data.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
