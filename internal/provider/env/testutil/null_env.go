package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// EnvResource is the subset of resource.Resource the null-env assertions drive.
type EnvResource interface {
	Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	Read(context.Context, resource.ReadRequest, *resource.ReadResponse)
	Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse)
}

// NullEnvClient replies with a null env and no SDK error, which IsNotFoundError cannot recognise as gone.
func NullEnvClient(t *testing.T, envField string) *client.Client {
	t.Helper()

	return newClient(t, func(string) string {
		return fmt.Sprintf(`{"data":{%q:null}}`, envField)
	})
}

// VanishingEnvClient serves a live env until the pending-MFA delete, then null envs: an env that vanishes mid-poll.
func VanishingEnvClient(t *testing.T, envField, deleteField string) *client.Client {
	t.Helper()
	shortenDeletionWait(t)

	deleted := false
	return newClient(t, func(operation string) string {
		if strings.HasPrefix(operation, "Delete") {
			deleted = true
			return fmt.Sprintf(`{"data":{%q:{"mutationId":"m","pendingMFA":true}}}`, deleteField)
		}
		if deleted {
			return fmt.Sprintf(`{"data":{%q:null}}`, envField)
		}
		return fmt.Sprintf(`{"data":{%q:{"name":"env","specRevision":1,"status":{}}}}`, envField)
	})
}

// shortenDeletionWait makes a delete stuck in PENDING_MFA fail in milliseconds instead of after the 5 minute MFA timeout.
func shortenDeletionWait(t *testing.T) {
	mfaTimeout, pollInterval := common.MFATimeout, common.DeletePollInterval
	common.MFATimeout, common.DeletePollInterval = time.Millisecond, time.Millisecond
	t.Cleanup(func() {
		common.MFATimeout, common.DeletePollInterval = mfaTimeout, pollInterval
	})
}

// newClient serializes body calls, so body may keep state without its own lock.
func newClient(t *testing.T, body func(operation string) string) *client.Client {
	t.Helper()

	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OperationName string `json:"operationName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		resp := body(req.OperationName)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, resp)
	}))
	t.Cleanup(srv.Close)

	return client.NewClient(srv.Client(), srv.URL, nil)
}

// envState nulls every attribute except the given values, enough for Read and Delete to reach their API calls.
func envState(t *testing.T, r EnvResource, values map[string]interface{}) tfsdk.State {
	t.Helper()
	ctx := context.Background()

	sresp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, sresp)

	objType, ok := sresp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("schema type is not a tftypes.Object")
	}

	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, values[name])
	}

	return tfsdk.State{Schema: sresp.Schema, Raw: tftypes.NewValue(objType, vals)}
}

// AssertReadRemovesNullEnv requires Read to drop the resource from state instead of panicking.
func AssertReadRemovesNullEnv(t *testing.T, r EnvResource) {
	t.Helper()
	ctx := context.Background()

	state := envState(t, r, map[string]interface{}{"name": "env", "id": "env"})

	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read errored: %v", resp.Diagnostics.Errors())
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected Read to remove the resource from state")
	}
}

// AssertDeleteSucceeds requires Delete to finish clean; reporting "not pending delete" instead of "gone" would strand a pending-MFA delete until its timeout.
func AssertDeleteSucceeds(t *testing.T, r EnvResource, forceDestroy bool) {
	t.Helper()

	if diags := runDelete(t, r, forceDestroy); diags.HasError() {
		t.Fatalf("Delete errored: %v", diags.Errors())
	}
}

// AssertDeleteLocked requires force_destroy to keep guarding an env that still exists.
func AssertDeleteLocked(t *testing.T, r EnvResource) {
	t.Helper()

	diags := runDelete(t, r, false)
	if !diags.HasError() || diags.Errors()[0].Summary() != "Env Locked" {
		t.Fatalf("expected Env Locked, got: %v", diags.Errors())
	}
}

func runDelete(t *testing.T, r EnvResource, forceDestroy bool) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()

	state := envState(t, r, map[string]interface{}{"name": "env", "id": "env", "force_destroy": forceDestroy})

	resp := &resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, resp)
	return resp.Diagnostics
}
