package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

func newAzureResource(c *client.Client) *AzureEnvResource {
	r := &AzureEnvResource{}
	r.Client = c
	return r
}

func TestAzureReadRemovesNullEnv(t *testing.T) {
	testutil.AssertReadRemovesNullEnv(t, newAzureResource(testutil.NullEnvClient(t, "azureEnv")))
}

func TestAzureDeleteAcceptsNullEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newAzureResource(testutil.NullEnvClient(t, "azureEnv")))
}

func TestAzureDeleteAcceptsVanishingEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newAzureResource(testutil.VanishingEnvClient(t, "azureEnv", "deleteAzureEnv")))
}
