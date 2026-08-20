package env

import (
	"context"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestEnvDataSourceBaseConfigure(t *testing.T) {
	base := &EnvDataSourceBase{}
	providerData := &sdk.AltinityCloudSDK{Client: &client.Client{}}

	resp := &datasource.ConfigureResponse{}
	base.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: providerData}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure errored: %v", resp.Diagnostics.Errors())
	}
	if base.Client != providerData.Client {
		t.Error("expected the client to be taken from the provider data")
	}
}

func TestEnvDataSourceBaseConfigureWithoutProviderData(t *testing.T) {
	base := &EnvDataSourceBase{}

	resp := &datasource.ConfigureResponse{}
	base.Configure(context.Background(), datasource.ConfigureRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no diagnostics, got: %v", resp.Diagnostics.Errors())
	}
	if base.Client != nil {
		t.Error("expected the client to stay unset")
	}
}
