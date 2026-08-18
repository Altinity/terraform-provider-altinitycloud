package hosted_env

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/testutil"
)

func TestHostedAWSModifyPlanSpecRevision(t *testing.T) {
	testutil.AssertModifyPlanSpecRevisionWithAttr(t, &HostedAWSEnvResource{}, "kms_key_arn")
}
