package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

func newHCloudResource(c *client.Client) *HCloudEnvResource {
	r := &HCloudEnvResource{}
	r.Client = c
	return r
}

func TestHCloudReadRemovesNullEnv(t *testing.T) {
	testutil.AssertReadRemovesNullEnv(t, newHCloudResource(testutil.NullEnvClient(t, "hcloudEnv")))
}

func TestHCloudDeleteAcceptsNullEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newHCloudResource(testutil.NullEnvClient(t, "hcloudEnv")))
}

func TestHCloudDeleteAcceptsVanishingEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newHCloudResource(testutil.VanishingEnvClient(t, "hcloudEnv", "deleteHCloudEnv")))
}
