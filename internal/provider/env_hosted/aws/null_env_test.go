package hosted_env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

func newAWSHostedResource(c *client.Client) *AWSEnvHostedResource {
	r := &AWSEnvHostedResource{}
	r.Client = c
	return r
}

func TestAWSHostedReadRemovesNullEnv(t *testing.T) {
	testutil.AssertReadRemovesNullEnv(t, newAWSHostedResource(testutil.NullEnvClient(t, "awsEnvHosted")))
}

func TestAWSHostedDeleteAcceptsNullEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newAWSHostedResource(testutil.NullEnvClient(t, "awsEnvHosted")), false)
}

func TestAWSHostedDeleteAcceptsVanishingEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newAWSHostedResource(testutil.VanishingEnvClient(t, "awsEnvHosted", "deleteAWSEnvHosted")), true)
}

func TestAWSHostedDeleteLocksLiveEnv(t *testing.T) {
	testutil.AssertDeleteLocked(t, newAWSHostedResource(testutil.VanishingEnvClient(t, "awsEnvHosted", "deleteAWSEnvHosted")))
}
