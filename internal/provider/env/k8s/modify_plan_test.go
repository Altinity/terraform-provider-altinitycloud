package env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
)

func TestK8SModifyPlanSpecRevision(t *testing.T) {
	testutil.AssertModifyPlanSpecRevision(t, &K8SEnvResource{})
}
