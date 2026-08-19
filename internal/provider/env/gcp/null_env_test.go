package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

func newGCPResource(c *client.Client) *GCPEnvResource {
	r := &GCPEnvResource{}
	r.Client = c
	return r
}

func TestGCPReadRemovesNullEnv(t *testing.T) {
	testutil.AssertReadRemovesNullEnv(t, newGCPResource(testutil.NullEnvClient(t, "gcpEnv")))
}

func TestGCPDeleteAcceptsNullEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newGCPResource(testutil.NullEnvClient(t, "gcpEnv")))
}

func TestGCPDeleteAcceptsVanishingEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newGCPResource(testutil.VanishingEnvClient(t, "gcpEnv", "deleteGCPEnv")))
}
