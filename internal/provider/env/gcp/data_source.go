package env

import (
	"context"
	"fmt"

	support "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &GCPEnvDataSource{}
	_ datasource.DataSourceWithConfigure = &GCPEnvDataSource{}
)

func NewGCPEnvDataSource() datasource.DataSource {
	return &GCPEnvDataSource{}
}

type GCPEnvDataSource struct {
	client *client.Client
}

func (d *GCPEnvDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_gcp"
}

func (d *GCPEnvDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GCPEnvDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Trace(ctx, "reading aws env state source")

	var data GCPEnvResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envName := data.Name.ValueString()
	apiResp, err := d.client.GetGCPEnv(ctx, envName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", support.FormatClientError(fmt.Sprintf("Unable to read env %s, got error: %s", envName, client.FormatError(err, envName))))
		return
	}

	if apiResp.GCPEnv == nil {
		resp.Diagnostics.AddError("Client Error", support.FormatClientError(fmt.Sprintf("Environment %s was not found", envName)))
		return
	}

	diags = data.toModel(*apiResp.GCPEnv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.Name

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
