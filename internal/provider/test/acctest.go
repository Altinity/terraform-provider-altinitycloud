package test

import (
	"os"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"altinitycloud": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("ALTINITYCLOUD_API_TOKEN"); v == "" {
		t.Fatal("ALTINITYCLOUD_API_TOKEN must be set for acceptance tests")
	}
}
