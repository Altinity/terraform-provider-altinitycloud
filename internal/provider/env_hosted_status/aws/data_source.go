package hosted_env_status

import (
	"context"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &AWSEnvHostedStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &AWSEnvHostedStatusDataSource{}

func NewAWSEnvHostedStatusDataSource() datasource.DataSource {
	return &AWSEnvHostedStatusDataSource{}
}

type AWSEnvHostedStatusDataSource struct {
	common.EnvStatusDataSourceBase
}

func (d *AWSEnvHostedStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_aws_hosted_status"
}

func (d *AWSEnvHostedStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading hosted aws env status data source")

	var data AWSEnvHostedStatusModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Read(ctx, common.MATCH_SPEC_TIMEOUT)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	refresh := func(ctx context.Context) (*common.PollResult, error) {
		apiResp, err := d.Client.GetAWSEnvHostedStatus(ctx, envName)
		if err != nil {
			return nil, err
		}
		if apiResp.AWSEnvHosted == nil {
			return &common.PollResult{}, nil
		}

		data.toModel(*apiResp.AWSEnvHosted)

		var errors []common.EnvError
		for _, e := range apiResp.AWSEnvHosted.Status.Errors {
			errors = append(errors, common.EnvError{Code: string(e.Code), Message: e.Message})
		}

		return &common.PollResult{
			AppliedSpecRevision: apiResp.AWSEnvHosted.Status.AppliedSpecRevision,
			Errors:              errors,
			Found:               true,
		}, nil
	}

	if !common.ReadEnvStatus(ctx, envName, data.WaitForAppliedSpecRevision.ValueInt64(), data.Verbose.ValueBool(), readTimeout, refresh, &resp.Diagnostics) {
		return
	}

	data.Id = data.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
