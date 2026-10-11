package env

import (
	"context"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

type EnvDataSourceBase struct {
	Client *client.Client
}

func (d *EnvDataSourceBase) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	sdk, ok := clientsupport.SDKFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}

	d.Client = sdk.Client
}
