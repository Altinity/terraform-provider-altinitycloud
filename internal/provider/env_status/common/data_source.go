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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
	if readTimeout == 0 {
		readTimeout = MATCH_SPEC_TIMEOUT
	}

	prefix := envName

	var tty *ttyWriter
	if verbose {
		tty = newTTYWriter()
		if tty != nil {
			defer tty.Close()
			tty.printf("%s: waiting for %q to be ready...\n", prefix, envName)
		}
	}

	start := time.Now()
	stateConf := &retry.StateChangeConf{
		Pending: []string{"WAITING", "CONNECTING"},
		Target:  []string{"READY"},
		Refresh: func() (interface{}, string, error) {
			result, err := poll(ctx, envName)
			if err != nil {
				return nil, "", fmt.Errorf("unable to read env status %s, got error: %s", envName, client.FormatError(err, envName))
			}

			if !result.Found {
				return nil, "", fmt.Errorf("environment %s was not found", envName)
			}

			elapsed := time.Since(start).Round(time.Second)

			if len(result.Errors) > 0 {
				var nonDisconnected []EnvError
				for _, e := range result.Errors {
					if e.Code == "DISCONNECTED" || e.Code == "K8S_DISCONNECTED" {
						if result.AppliedSpecRevision == 0 {
							if tty != nil {
								tty.printf("%s: [%s] waiting for initial connection (not yet provisioned)...\n",
									prefix, elapsed)
							}
							continue
						}
						nonDisconnected = append(nonDisconnected, e)
						continue
					}
					nonDisconnected = append(nonDisconnected, e)
				}

				if len(nonDisconnected) == 0 {
					if tty != nil {
						tty.printf("%s: [%s] connecting...\n", prefix, elapsed)
					}
					return result, "CONNECTING", nil
				}

				var errorDetails string
				for _, e := range nonDisconnected {
					errorDetails += fmt.Sprintf("%s: %s\n", e.Code, e.Message)
				}
				return nil, "", fmt.Errorf("environment %s has provisioning errors:\n%s", envName, errorDetails)
			}

			if result.AppliedSpecRevision >= targetRevision {
				if tty != nil {
					tty.printf("%s: [%s] ready!\n", prefix, elapsed)
				}
				return result, "READY", nil
			}

			if tty != nil {
				tty.printf("%s: [%s] waiting...\n", prefix, elapsed)
			}
			return result, "WAITING", nil
		},
		Timeout:      readTimeout,
		PollInterval: MATCH_SPEC_POLL_INTERVAL,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		diags.AddError("Status Error", fmt.Sprintf("Error waiting for env status %s: %s", envName, err))
		return false
	}
	return true
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
