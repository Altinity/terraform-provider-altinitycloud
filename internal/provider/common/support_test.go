package common

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestAddClientError_AppendsSupportMessage(t *testing.T) {
	var diags diag.Diagnostics

	AddClientError(&diags, "something failed")

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Summary() != "Client Error" {
		t.Errorf("expected summary %q, got %q", "Client Error", diags[0].Summary())
	}
	if !strings.Contains(diags[0].Detail(), "something failed") {
		t.Errorf("detail does not contain original message: %q", diags[0].Detail())
	}
	if !strings.Contains(diags[0].Detail(), "Slack") {
		t.Errorf("detail does not contain support message: %q", diags[0].Detail())
	}
}

func TestAddClientError_DeduplicatesSupportMessage(t *testing.T) {
	var diags diag.Diagnostics

	AddClientError(&diags, "first error")
	AddClientError(&diags, "second error")
	AddClientError(&diags, "third error")

	if len(diags) != 3 {
		t.Fatalf("expected 3 diagnostics, got %d", len(diags))
	}

	total := 0
	for _, d := range diags {
		total += strings.Count(d.Detail(), supportMessage)
	}
	if total != 1 {
		t.Errorf("expected support message to appear exactly once across diagnostics, got %d", total)
	}

	if !strings.Contains(diags[0].Detail(), supportMessage) {
		t.Errorf("first diagnostic should carry the support message")
	}
	if strings.Contains(diags[1].Detail(), supportMessage) {
		t.Errorf("second diagnostic should not carry the support message")
	}
	if strings.Contains(diags[2].Detail(), supportMessage) {
		t.Errorf("third diagnostic should not carry the support message")
	}
}

func TestAddSupportError_UsesProvidedSummary(t *testing.T) {
	var diags diag.Diagnostics

	AddSupportError(&diags, "Delete Error", "delete failed")

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Summary() != "Delete Error" {
		t.Errorf("expected summary %q, got %q", "Delete Error", diags[0].Summary())
	}
	if !strings.Contains(diags[0].Detail(), "Slack") {
		t.Errorf("detail does not contain support message: %q", diags[0].Detail())
	}
}

func TestAddSupportError_DedupAcrossDifferentSummaries(t *testing.T) {
	var diags diag.Diagnostics

	AddClientError(&diags, "client failure")
	AddSupportError(&diags, "Delete Error", "delete failure")
	AddSupportError(&diags, "Status Error", "status failure")

	total := 0
	for _, d := range diags {
		total += strings.Count(d.Detail(), supportMessage)
	}
	if total != 1 {
		t.Errorf("expected support message to appear exactly once across all error types, got %d", total)
	}
}

func TestAddClientError_EmptyDetail(t *testing.T) {
	var diags diag.Diagnostics

	AddClientError(&diags, "")

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !strings.Contains(diags[0].Detail(), "Slack") {
		t.Errorf("detail should contain support message even for empty input")
	}
}
