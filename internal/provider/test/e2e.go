//go:build e2e

package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
