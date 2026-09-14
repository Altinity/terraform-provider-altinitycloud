package common

import (
	"context"
	"fmt"
	"time"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StatusRefreshFunc fetches the current env status and writes it into the
// caller's model. It reports Found=false when the query succeeded but returned
// no environment, which carries no SDK error to inspect.
type StatusRefreshFunc func(ctx context.Context) (*PollResult, error)

// ReadEnvStatus runs the env status read shared by every provider: fetch once,
// return early when the applied spec revision already satisfies the target, and
// otherwise wait for it and refresh so the model carries the settled status.
// Returns false when diagnostics were added and the caller should stop.
func ReadEnvStatus(ctx context.Context, envName string, targetRevision int64, verbose bool, readTimeout time.Duration, refresh StatusRefreshFunc, diags *diag.Diagnostics) bool {
	result, ok := refreshStatus(ctx, envName, refresh, diags)
	if !ok {
		return false
	}

	if targetRevision == 0 || result.AppliedSpecRevision >= targetRevision {
		tflog.Trace(ctx, "env status matches spec", map[string]interface{}{"name": envName})
		return true
	}

	if !waitForSpecRevision(ctx, envName, targetRevision, verbose, refresh, diags, readTimeout) {
		return false
	}

	// Refresh again so the model reflects the status that satisfied the wait.
	_, ok = refreshStatus(ctx, envName, refresh, diags)
	return ok
}

func refreshStatus(ctx context.Context, envName string, refresh StatusRefreshFunc, diags *diag.Diagnostics) (*PollResult, bool) {
	result, err := refresh(ctx)
	if err != nil {
		clientsupport.AddClientError(diags, fmt.Sprintf("Unable to read env status %s, got error: %s", envName, client.FormatError(err, envName)))
		return nil, false
	}

	if !result.Found {
		clientsupport.AddClientError(diags, fmt.Sprintf("Environment %s was not found", envName))
		return nil, false
	}

	return result, true
}
