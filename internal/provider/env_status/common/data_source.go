package common

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var MATCH_SPEC_TIMEOUT = time.Duration(60) * time.Minute
var MATCH_SPEC_POLL_INTERVAL = 30 * time.Second

type EnvStatusDataSourceBase struct {
	Client *client.Client
}

func (d *EnvStatusDataSourceBase) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*sdk.AltinityCloudSDK)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.AltinityCloudSDK, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.Client = sdk.Client
}

// EnvError represents a provisioning error from the API.
type EnvError struct {
	Code    string
	Message string
}

// PollResult holds the relevant fields extracted from a provider-specific API response.
type PollResult struct {
	AppliedSpecRevision int64
	Errors              []EnvError
	Found               bool
}

// PollFunc is a callback that fetches the current env status from the API.
type PollFunc func(ctx context.Context, envName string) (*PollResult, error)

// WaitForSpecRevision polls the environment status until the applied spec revision
// matches the target revision. It handles TTY output, DISCONNECTED errors, and timeouts.
// Returns true if the target revision was reached, false otherwise (errors added to diags).
func WaitForSpecRevision(ctx context.Context, envName string, targetRevision int64, verbose bool, poll PollFunc, diags *diag.Diagnostics, readTimeout time.Duration) bool {
	var tty *ttyWriter
	if verbose {
		tty = newTTYWriter()
		if tty != nil {
			defer tty.Close()
			tty.printf("[altinitycloud] Waiting for %q to reach spec revision %d...\n", envName, targetRevision)
		}
	}

	if readTimeout == 0 {
		readTimeout = MATCH_SPEC_TIMEOUT
	}

	start := time.Now()
	timeout := time.After(readTimeout)
	ticker := time.NewTicker(MATCH_SPEC_POLL_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			diags.AddError("Context Cancelled", "The context was cancelled, stopping env status read.")
			return false
		case <-timeout:
			diags.AddError("Timeout", "Timeout reached while waiting for env satus to match spec.")
			return false
		case <-ticker.C:
			tflog.Trace(ctx, "checking if env match spec", map[string]interface{}{"name": envName})
			result, err := poll(ctx, envName)

			if err != nil {
				diags.AddError("Client Error", fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
				return false
			}

			if !result.Found {
				diags.AddError("Client Error", fmt.Sprintf("Environment %s was not found", envName))
				return false
			}

			elapsed := time.Since(start).Round(time.Second)
			errorCount := len(result.Errors)
			if errorCount > 0 {
				// Check if the only error is DISCONNECTED (transient, keep polling)
				if errorCount == 1 && result.Errors[0].Code == "DISCONNECTED" {
					if tty != nil {
						tty.printf("[altinitycloud] [%s] %s: revision %d/%d, connecting...\n",
							elapsed, envName, result.AppliedSpecRevision, targetRevision)
					}
					continue
				}

				var errorDetails string
				for _, e := range result.Errors {
					errorDetails += fmt.Sprintf("%s: %s\n", e.Code, e.Message)
				}
				diags.AddError("Client Error", fmt.Sprintf("Environment %s has provisioning errors:\n%s", envName, errorDetails))
				return false
			}

			if result.AppliedSpecRevision >= targetRevision {
				tflog.Trace(ctx, "env status matchs spec", map[string]interface{}{"name": envName})
				if tty != nil {
					tty.printf("[altinitycloud] [%s] %s: revision %d/%d, ready!\n",
						elapsed, envName, result.AppliedSpecRevision, targetRevision)
				}
				return true
			}

			if tty != nil {
				tty.printf("[altinitycloud] [%s] %s: revision %d/%d, waiting...\n",
					elapsed, envName, result.AppliedSpecRevision, targetRevision)
			}
		}
	}
}

// ttyWriter wraps a file handle to /dev/tty for real-time progress output.
type ttyWriter struct {
	file *os.File
}

// newTTYWriter attempts to open the controlling terminal directly.
// Returns nil if no TTY is available (e.g. CI/CD environments).
func newTTYWriter() *ttyWriter {
	f, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return nil
	}
	return &ttyWriter{file: f}
}

func (w *ttyWriter) printf(format string, args ...interface{}) {
	fmt.Fprintf(w.file, format, args...)
}

func (w *ttyWriter) Close() {
	w.file.Close()
}
