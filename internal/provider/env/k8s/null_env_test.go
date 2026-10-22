package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

func newK8SResource(c *client.Client) *K8SEnvResource {
	r := &K8SEnvResource{}
	r.Client = c
	return r
}

func TestK8SReadRemovesNullEnv(t *testing.T) {
	testutil.AssertReadRemovesNullEnv(t, newK8SResource(testutil.NullEnvClient(t, "k8sEnv")))
}

func TestK8SDeleteAcceptsNullEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newK8SResource(testutil.NullEnvClient(t, "k8sEnv")))
}

func TestK8SDeleteAcceptsVanishingEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newK8SResource(testutil.VanishingEnvClient(t, "k8sEnv", "deleteK8SEnv")))
}
