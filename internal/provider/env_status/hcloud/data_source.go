package env_status

import (
	"context"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &HCloudEnvStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &HCloudEnvStatusDataSource{}

func NewHCloudEnvStatusDataSource() datasource.DataSource {
	return &HCloudEnvStatusDataSource{}
}

type HCloudEnvStatusDataSource struct {
	common.EnvStatusDataSourceBase
}

func (d *HCloudEnvStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_hcloud_status"
}

func (d *HCloudEnvStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading hcloud env status data source")

	var data HCloudEnvStatusModel
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
		apiResp, err := d.Client.GetHCloudEnvStatus(ctx, envName)
		if err != nil {
			return nil, err
		}
		if apiResp.HcloudEnv == nil {
			return &common.PollResult{}, nil
		}

		data.toModel(*apiResp.HcloudEnv)

		var errors []common.EnvError
		for _, e := range apiResp.HcloudEnv.Status.Errors {
			errors = append(errors, common.EnvError{Code: string(e.Code), Message: e.Message})
		}

		return &common.PollResult{
			AppliedSpecRevision: apiResp.HcloudEnv.Status.AppliedSpecRevision,
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
