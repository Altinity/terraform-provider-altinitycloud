package env_status

import (
	"context"
	"fmt"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
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
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.Client.GetK8SEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.K8sEnv == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	waitForAppliedSpecRevision := data.WaitForAppliedSpecRevision.ValueInt64()
	if waitForAppliedSpecRevision == 0 || apiResp.K8sEnv.Status.AppliedSpecRevision >= waitForAppliedSpecRevision {
		tflog.Trace(ctx, "env status matchs spec", map[string]interface{}{"name": envName})
		data.toModel(*apiResp.K8sEnv)
		data.Id = data.Name

		diags = resp.State.Set(ctx, &data)
		resp.Diagnostics.Append(diags...)

		return
	}

	poll := func(ctx context.Context, name string) (*common.PollResult, error) {
		resp, err := d.Client.GetK8SEnvStatus(ctx, name)
		if err != nil {
			return nil, err
		}
		if resp.K8sEnv == nil {
			return &common.PollResult{Found: false}, nil
		}
		var errors []common.EnvError
		for _, e := range resp.K8sEnv.Status.Errors {
			errors = append(errors, common.EnvError{Code: string(e.Code), Message: e.Message})
		}
		return &common.PollResult{
			AppliedSpecRevision: resp.K8sEnv.Status.AppliedSpecRevision,
			Errors:              errors,
			Found:               true,
		}, nil
	}

	if !common.WaitForSpecRevision(ctx, envName, waitForAppliedSpecRevision, data.Verbose.ValueBool(), poll, &resp.Diagnostics) {
		return
	}

	apiResp, err = d.Client.GetK8SEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	data.toModel(*apiResp.K8sEnv)
	data.Id = data.Name
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
