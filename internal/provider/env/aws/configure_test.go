package env

import (
	"context"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	sdkclient "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// The bases are embedded, so Configure reaches the concrete type by method
// promotion. A base held by value with a pointer receiver would silently leave
// the client unset if that ever stopped holding.
func TestAWSDataSourceConfigurePromotesToBase(t *testing.T) {
	d := &AWSEnvDataSource{}
	providerData := &sdk.AltinityCloudSDK{Client: &sdkclient.Client{}}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: providerData}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure errored: %v", resp.Diagnostics.Errors())
	}
	if d.Client != providerData.Client {
		t.Error("expected the embedded base to receive the client")
	}
}

func TestAWSResourceConfigurePromotesToBase(t *testing.T) {
	r := &AWSEnvResource{}
	providerData := &sdk.AltinityCloudSDK{Client: &sdkclient.Client{}}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: providerData}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure errored: %v", resp.Diagnostics.Errors())
	}
	if r.Client != providerData.Client {
		t.Error("expected the embedded base to receive the client")
	}
}
