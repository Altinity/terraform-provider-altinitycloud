//go:build e2e

package env_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
)

// TestE2EAltinityCloudEnvAWS drives create -> update against the dev control
// plane (https://anywhere.dev.altinity.cloud) using a dummy-prefixed env, and
// asserts there is no drift after each apply (a non-empty plan = drift).
//
// Teardown is intentionally skipped: deleting an env on dev requires MFA
// confirmation, which CI cannot provide. Dummy-prefixed envs are cleaned up out
// of band.
func TestE2EAltinityCloudEnvAWS(t *testing.T) {
	test.E2EPreCheck(t)
	tf, workdir := test.NewE2ETerraform(t)
	ctx := context.Background()

	envName := "dummy-e2e-aws-" + test.GenerateRandomResourceName()

	// 1. Create.
	test.WriteE2EConfig(t, workdir, awsE2EConfig(envName, 1))
	if err := tf.Apply(ctx); err != nil {
		t.Fatalf("create apply failed: %s", err)
	}

	// 2. Drift check: a plan right after create must be empty.
	if changed, err := tf.Plan(ctx); err != nil {
		t.Fatalf("plan after create failed: %s", err)
	} else if changed {
		t.Fatalf("unexpected drift after create for env %s", envName)
	}

	// 3. Update (capacity 1 -> 3), then assert no drift.
	test.WriteE2EConfig(t, workdir, awsE2EConfig(envName, 3))
	if err := tf.Apply(ctx); err != nil {
		t.Fatalf("update apply failed: %s", err)
	}
	if changed, err := tf.Plan(ctx); err != nil {
		t.Fatalf("plan after update failed: %s", err)
	} else if changed {
		t.Fatalf("unexpected drift after update for env %s", envName)
	}

	t.Logf("e2e create+update+drift OK for %s (delete skipped: dev requires MFA)", envName)
}

func awsE2EConfig(envName string, capacity int) string {
	return fmt.Sprintf(`
resource "%s" "dummy" {
  name           = "%s"
  cidr           = "10.0.0.0/16"
  region         = "us-east-1"
  aws_account_id = "123456789012"
  zones          = ["us-east-1a", "us-east-1b"]

  node_groups = [{
    zones             = ["us-east-1a"]
    node_type         = "t4g.large"
    capacity_per_zone = %d
    reservations      = ["SYSTEM", "CLICKHOUSE", "ZOOKEEPER"]
  }]

  force_destroy                   = true
  skip_deprovision_on_destroy     = true
  allow_delete_while_disconnected = true
}
`, RESOURCE_NAME, envName, capacity)
}
