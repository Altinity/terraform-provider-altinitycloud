package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

func newAWSResource(c *client.Client) *AWSEnvResource {
	r := &AWSEnvResource{}
	r.Client = c
	return r
}

func TestAWSReadRemovesNullEnv(t *testing.T) {
	testutil.AssertReadRemovesNullEnv(t, newAWSResource(testutil.NullEnvClient(t, "awsEnv")))
}

func TestAWSDeleteAcceptsNullEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newAWSResource(testutil.NullEnvClient(t, "awsEnv")))
}

func TestAWSDeleteAcceptsVanishingEnv(t *testing.T) {
	testutil.AssertDeleteSucceeds(t, newAWSResource(testutil.VanishingEnvClient(t, "awsEnv", "deleteAWSEnv")))
}
