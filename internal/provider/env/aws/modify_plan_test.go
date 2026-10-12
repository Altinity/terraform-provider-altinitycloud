package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
)

func TestAWSModifyPlanSpecRevision(t *testing.T) {
	testutil.AssertModifyPlanSpecRevision(t, &AWSEnvResource{})
}

func TestAWSModifyPlanMFA(t *testing.T) {
	testutil.AssertModifyPlanMFA(t, &AWSEnvResource{})
}
