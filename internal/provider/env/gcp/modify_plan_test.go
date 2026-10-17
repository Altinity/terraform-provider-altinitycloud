package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
)

func TestGCPModifyPlanSpecRevision(t *testing.T) {
	testutil.AssertModifyPlanSpecRevision(t, &GCPEnvResource{})
}

func TestGCPModifyPlanMFA(t *testing.T) {
	testutil.AssertModifyPlanMFA(t, &GCPEnvResource{})
}
