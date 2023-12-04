package env

import (
	"context"
	"fmt"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &AWSEnvDataSource{}
	_ datasource.DataSourceWithConfigure = &AWSEnvDataSource{}
)

func NewAWSEnvDataSource() datasource.DataSource {
	return &AWSEnvDataSource{}
}

type AWSEnvDataSource struct {
	client *client.Client
}

func (d *AWSEnvDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_aws"
}

func (d *AWSEnvDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*sdk.AltinityCloudSDK)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = sdk.Client
}

func (d *AWSEnvDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading aws env state source")

	var data AWSEnvResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.client.GetAWSEnv(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read env %s, got error: %s", envName, err))
		return
	}

	if apiResp.AwsEnv == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
		return
	}

	data.toModel(*apiResp.AwsEnv)
	data.Id = data.Name

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
