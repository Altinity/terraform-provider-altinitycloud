//go:build e2e

package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// E2EAPIURL is the dev control plane the e2e tests talk to. Envs are created
// with a "dummy-" name prefix, which the dev control plane treats as a sandbox
// (no real cloud provisioning).
const E2EAPIURL = "https://anywhere.dev.altinity.cloud"

const e2eProviderSource = "altinity/altinitycloud"

// E2EPreCheck skips the test when no API token is available (local runs, or PRs
// from forks where the secret is not exposed).
func E2EPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("ALTINITYCLOUD_API_TOKEN") == "" {
		t.Skip("ALTINITYCLOUD_API_TOKEN not set; skipping e2e tests against dev control plane")
	}
}

// NewE2ETerraform builds the provider, wires a dev_overrides CLI config so a
// real Terraform binary uses that local build (no init/registry), points the
// provider at the dev control plane, and returns a tfexec handle plus its
// working directory.
//
// These tests drive Terraform directly instead of using the plugin-testing
// harness because that harness always runs a post-test destroy, and deleting an
// env on the dev control plane requires MFA confirmation that CI cannot provide.
// We therefore exercise create/update/drift and intentionally skip teardown;
// dummy-prefixed envs are cleaned up out of band.
func NewE2ETerraform(t *testing.T) (*tfexec.Terraform, string) {
	t.Helper()

	tfPath, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform binary not found in PATH; skipping e2e tests")
	}

	pluginDir := t.TempDir()
	bin := filepath.Join(pluginDir, "terraform-provider-altinitycloud")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/altinity/terraform-provider-altinitycloud").CombinedOutput(); err != nil {
		t.Fatalf("failed to build provider: %s\n%s", err, out)
	}

	cliConfig := filepath.Join(t.TempDir(), "dev.tfrc")
	cliBody := fmt.Sprintf(`provider_installation {
  dev_overrides {
    %q = %q
  }
  direct {}
}
`, e2eProviderSource, pluginDir)
	if err := os.WriteFile(cliConfig, []byte(cliBody), 0o600); err != nil {
		t.Fatalf("failed to write CLI config: %s", err)
	}

	t.Setenv("TF_CLI_CONFIG_FILE", cliConfig)
	t.Setenv("ALTINITYCLOUD_API_URL", E2EAPIURL)

	workdir := t.TempDir()
	tf, err := tfexec.NewTerraform(workdir, tfPath)
	if err != nil {
		t.Fatalf("failed to init tfexec: %s", err)
	}
	return tf, workdir
}

// e2eTransientPatterns are control-plane/network errors worth retrying in e2e.
// The dev control plane intermittently resets long-lived connections (~30s) on
// heavy create requests, which surfaces as a transient apply failure from the
// GitHub Actions runners.
var e2eTransientPatterns = []string{
	"connection reset by peer",
	"connection refused",
	"i/o timeout",
	"TLS handshake timeout",
	"EOF",
	"request failed",
	"502", "503", "504",
}

func isE2ETransient(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, p := range e2eTransientPatterns {
		if strings.Contains(msg, p) {
			return true
		}
	}
	return false
}

// isE2EAlreadyExists reports a name collision. The dev control plane can commit
// an env server-side and still surface the create as a failure (e.g. it resets
// a heavy create connection after committing), so a later attempt with the same
// name collides. Retrying with a fresh name recovers, same as a transient error.
func isE2EAlreadyExists(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

// RunE2ELifecycle drives the standard env lifecycle against the dev control
// plane: create -> drift check -> update -> drift check. Teardown is skipped
// (dev delete requires MFA). configFn renders the resource HCL for a given env
// name and node capacity (the capacity drives the mutable update).
//
// Create is retried with a fresh env name on transient dev errors so a reset
// connection doesn't leave us colliding with a half-created env; the idempotent
// update/plan steps are retried in place.
func RunE2ELifecycle(t *testing.T, namePrefix string, configFn func(envName string, capacity int) string) {
	t.Helper()
	E2EPreCheck(t)
	tf, workdir := NewE2ETerraform(t)
	ctx := context.Background()

	const maxAttempts = 3

	var envName string
	for attempt := 1; ; attempt++ {
		envName = namePrefix + GenerateRandomResourceName()
		WriteE2EConfig(t, workdir, configFn(envName, 1))
		err := tf.Apply(ctx)
		if err == nil {
			break
		}
		retryable := isE2ETransient(err) || isE2EAlreadyExists(err)
		if attempt >= maxAttempts || !retryable {
			t.Fatalf("create apply failed: %s", err)
		}
		t.Logf("retryable create error (attempt %d/%d), retrying with fresh name: %s", attempt, maxAttempts, err)
	}

	e2eAssertNoDrift(t, tf, ctx, envName, "create")

	test := func() error {
		WriteE2EConfig(t, workdir, configFn(envName, 3))
		return tf.Apply(ctx)
	}
	if err := e2eRetryTransient(t, maxAttempts, test); err != nil {
		t.Fatalf("update apply failed: %s", err)
	}
	e2eAssertNoDrift(t, tf, ctx, envName, "update")

	t.Logf("e2e create+update+drift OK for %s (delete skipped: dev requires MFA)", envName)
}

func e2eAssertNoDrift(t *testing.T, tf *tfexec.Terraform, ctx context.Context, envName, phase string) {
	t.Helper()
	var changed bool
	err := e2eRetryTransient(t, 3, func() error {
		var planErr error
		changed, planErr = tf.Plan(ctx)
		return planErr
	})
	if err != nil {
		t.Fatalf("plan after %s failed: %s", phase, err)
	}
	if changed {
		t.Fatalf("unexpected drift after %s for env %s", phase, envName)
	}
}

func e2eRetryTransient(t *testing.T, attempts int, fn func() error) error {
	t.Helper()
	var err error
	for i := 1; i <= attempts; i++ {
		err = fn()
		if err == nil || !isE2ETransient(err) {
			return err
		}
		t.Logf("transient error (attempt %d/%d): %s", i, attempts, err)
	}
	return err
}

// WriteE2EConfig writes main.tf (terraform block + given resource HCL) into the
// working directory.
func WriteE2EConfig(t *testing.T, workdir, resourceHCL string) {
	t.Helper()
	hcl := `terraform {
  required_providers {
    altinitycloud = {
      source = "` + e2eProviderSource + `"
    }
  }
}

` + resourceHCL
	if err := os.WriteFile(filepath.Join(workdir, "main.tf"), []byte(hcl), 0o600); err != nil {
		t.Fatalf("failed to write main.tf: %s", err)
	}
}
