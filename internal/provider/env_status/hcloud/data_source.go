package env_status

import (
	"context"
	"fmt"
	"time"

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
	tflog.Trace(ctx, "reading aws env status data source")

	var data HCloudEnvStatusModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.Client.GetHCloudEnvStatus(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, err))
		return
	}

	if apiResp.HcloudEnv == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	waitForAppliedSpecRevision := data.WaitForAppliedSpecRevision.ValueInt64()
	if waitForAppliedSpecRevision == 0 || apiResp.HcloudEnv.Status.AppliedSpecRevision >= waitForAppliedSpecRevision {
		tflog.Trace(ctx, "env status matchs spec", map[string]interface{}{"name": envName})
		data.toModel(*apiResp.HcloudEnv)
		data.Id = data.Name

		diags = resp.State.Set(ctx, &data)
		resp.Diagnostics.Append(diags...)

		return
	}

	// Polling to wait for match spec to complete
	timeout := time.After(common.MATCH_SPEC_TIMEOUT)
	ticker := time.NewTicker(common.MATCH_SPEC_POLL_INTERVAL)
	defer ticker.Stop()

tickerLoop:
	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Context Cancelled", "The context was cancelled, stopping env status read.")
			return
		case <-timeout:
			if len(apiResp.HcloudEnv.Status.Errors) > 0 {
				var errorDetails string
				for _, err := range apiResp.HcloudEnv.Status.Errors {
					errorDetails += fmt.Sprintf("%s: %s\n", err.Code, err.Message)
				}
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s has provisioning errors:\n%s", envName, errorDetails))
				return
			}

			resp.Diagnostics.AddError("Timeout", "Timeout reached while waiting for env satus to match spec.")
			return
		case <-ticker.C:
			tflog.Trace(ctx, "checking if env match spec", map[string]interface{}{"name": envName})
			apiResp, err := d.Client.GetHCloudEnvStatus(ctx, envName)

			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, err))
				return
			}

			if apiResp.HcloudEnv == nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
				return
			}

			errorCount := len(apiResp.HcloudEnv.Status.Errors)
			if errorCount > 0 {
				var errorDetails string
				for _, err := range apiResp.HcloudEnv.Status.Errors {
					// Ignore DISCONNECTED errors since will interrup matching spec with new provisioning envs.
					if errorCount == 1 && err.Code == "DISCONNECTED" {
						continue tickerLoop
					}
					errorDetails += fmt.Sprintf("%s: %s\n", err.Code, err.Message)
				}
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s has provisioning errors:\n%s", envName, errorDetails))
				return
			}

			if apiResp.HcloudEnv.Status.AppliedSpecRevision >= waitForAppliedSpecRevision {
				tflog.Trace(ctx, "env status matchs spec", map[string]interface{}{"name": envName})
				data.toModel(*apiResp.HcloudEnv)
				data.Id = data.Name

				diags = resp.State.Set(ctx, &data)
				resp.Diagnostics.Append(diags...)

				return
			}
		}
	}
}
