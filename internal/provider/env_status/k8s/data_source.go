package env_status

import (
	"context"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &K8SEnvStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &K8SEnvStatusDataSource{}

func NewK8SEnvStatusDataSource() datasource.DataSource {
	return &K8SEnvStatusDataSource{}
}

type K8SEnvStatusDataSource struct {
	common.EnvStatusDataSourceBase
}

func (d *K8SEnvStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_k8s_status"
}

func (d *K8SEnvStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading k8s env status data source")

	var data K8SEnvStatusModel
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
		apiResp, err := d.Client.GetK8SEnvStatus(ctx, envName)
		if err != nil {
			return nil, err
		}
		if apiResp.K8sEnv == nil {
			return &common.PollResult{}, nil
		}

		data.toModel(*apiResp.K8sEnv)

		var errors []common.EnvError
		for _, e := range apiResp.K8sEnv.Status.Errors {
			errors = append(errors, common.EnvError{Code: string(e.Code), Message: e.Message})
		}

		return &common.PollResult{
			AppliedSpecRevision: apiResp.K8sEnv.Status.AppliedSpecRevision,
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
