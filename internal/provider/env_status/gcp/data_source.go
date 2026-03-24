package env_status

import (
	"context"
	"fmt"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_status/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &GCPEnvStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &GCPEnvStatusDataSource{}

func NewGCPEnvStatusDataSource() datasource.DataSource {
	return &GCPEnvStatusDataSource{}
}

type GCPEnvStatusDataSource struct {
	common.EnvStatusDataSourceBase
}

func (d *GCPEnvStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_gcp_status"
}

func (d *GCPEnvStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading gcp env status data source")

	var data GCPEnvStatusModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.Client.GetGCPEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.GCPEnv == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	waitForAppliedSpecRevision := data.WaitForAppliedSpecRevision.ValueInt64()
	if waitForAppliedSpecRevision == 0 || apiResp.GCPEnv.Status.AppliedSpecRevision >= waitForAppliedSpecRevision {
		tflog.Trace(ctx, "env status matchs spec", map[string]interface{}{"name": envName})
		data.toModel(*apiResp.GCPEnv)
		data.Id = data.Name

		diags = resp.State.Set(ctx, &data)
		resp.Diagnostics.Append(diags...)

		return
	}

	poll := func(ctx context.Context, name string) (*common.PollResult, error) {
		resp, err := d.Client.GetGCPEnvStatus(ctx, name)
		if err != nil {
			return nil, err
		}
		if resp.GCPEnv == nil {
			return &common.PollResult{Found: false}, nil
		}
		var errors []common.EnvError
		for _, e := range resp.GCPEnv.Status.Errors {
			errors = append(errors, common.EnvError{Code: string(e.Code), Message: e.Message})
		}
		return &common.PollResult{
			AppliedSpecRevision: resp.GCPEnv.Status.AppliedSpecRevision,
			Errors:              errors,
			Found:               true,
		}, nil
	}

	if !common.WaitForSpecRevision(ctx, envName, waitForAppliedSpecRevision, data.Verbose.ValueBool(), poll, &resp.Diagnostics) {
		return
	}

	apiResp, err = d.Client.GetGCPEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	data.toModel(*apiResp.GCPEnv)
	data.Id = data.Name
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
