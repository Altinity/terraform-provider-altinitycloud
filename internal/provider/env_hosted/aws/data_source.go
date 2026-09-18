package hosted_env

import (
	"context"
	"fmt"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &AWSEnvHostedDataSource{}
	_ datasource.DataSourceWithConfigure = &AWSEnvHostedDataSource{}
)

func NewAWSEnvHostedDataSource() datasource.DataSource {
	return &AWSEnvHostedDataSource{}
}

type AWSEnvHostedDataSource struct {
	common.EnvDataSourceBase
}

func (d *AWSEnvHostedDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_aws_hosted"
}

func (d *AWSEnvHostedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading hosted aws env data source")

	var data AWSEnvHostedResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.Client.GetAWSEnvHosted(ctx, envName)
	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.AWSEnvHosted == nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	resp.Diagnostics.Append(data.applySpec(ctx, apiResp.AWSEnvHosted.Name, apiResp.AWSEnvHosted.Spec, apiResp.AWSEnvHosted.SpecRevision)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Data sources are read-only, so expose custom domains exactly as returned.
	customDomains, diags := common.ListToModel(apiResp.AWSEnvHosted.Spec.CustomDomains)
	resp.Diagnostics.Append(diags...)
	data.CustomDomains = customDomains
	data.Id = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
