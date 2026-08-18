package hosted_env

import (
	"context"
	"fmt"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &HostedAWSEnvDataSource{}
	_ datasource.DataSourceWithConfigure = &HostedAWSEnvDataSource{}
)

func NewHostedAWSEnvDataSource() datasource.DataSource {
	return &HostedAWSEnvDataSource{}
}

type HostedAWSEnvDataSource struct {
	client *client.Client
}

func (d *HostedAWSEnvDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosted_env_aws"
}

func (d *HostedAWSEnvDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*sdk.AltinityCloudSDK)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.AltinityCloudSDK, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = sdk.Client
}

func (d *HostedAWSEnvDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading hosted aws env data source")

	var data HostedAWSEnvResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.client.GetHostedAWSEnv(ctx, envName)
	if err != nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Unable to read env %s, got error: %s", envName, client.FormatError(err, envName)))
		return
	}

	if apiResp.HostedAWSEnv == nil {
		clientsupport.AddClientError(&resp.Diagnostics, fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	resp.Diagnostics.Append(data.applySpec(ctx, apiResp.HostedAWSEnv.Name, apiResp.HostedAWSEnv.Spec, apiResp.HostedAWSEnv.SpecRevision)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Data sources are read-only, so expose custom domains exactly as returned.
	customDomains, diags := common.ListToModel(apiResp.HostedAWSEnv.Spec.CustomDomains)
	resp.Diagnostics.Append(diags...)
	data.CustomDomains = customDomains
	data.Id = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
