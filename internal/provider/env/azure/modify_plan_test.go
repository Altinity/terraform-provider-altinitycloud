package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
)

func TestAzureModifyPlanSpecRevision(t *testing.T) {
	testutil.AssertModifyPlanSpecRevision(t, &AzureEnvResource{})
}

func TestAzureModifyPlanMFA(t *testing.T) {
	testutil.AssertModifyPlanMFA(t, &AzureEnvResource{})
}
