package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPreviewBodyForError(t *testing.T) {
	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		if got := PreviewBodyForError(nil); got != "" {
			t.Fatalf("got %q want empty", got)
		}
	})
	t.Run("short unchanged", func(t *testing.T) {
		t.Parallel()
		in := []byte("short message")
		if got := PreviewBodyForError(in); got != "short message" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("newlines collapsed", func(t *testing.T) {
		t.Parallel()
		in := []byte("a\nb\rc")
		if got := PreviewBodyForError(in); got != "a b c" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("truncates long body", func(t *testing.T) {
		t.Parallel()
		in := make([]byte, maxErrorBodyPreview+100)
		for i := range in {
			in[i] = 'x'
		}
		got := PreviewBodyForError(in)
		if !strings.Contains(got, "truncated") {
			t.Fatalf("expected truncated note: %q", got)
		}
		if !strings.Contains(got, "bytes total") {
			t.Fatalf("expected byte count: %q", got)
		}
		if len(got) > 800 {
			t.Fatalf("error preview unexpectedly long: %d chars", len(got))
		}
	})
}

func TestSanitizeRequestURL(t *testing.T) {
	t.Parallel()
	raw := "https://api.example.com/v1/sign?token=secret&x=1#frag"
	got := SanitizeRequestURL(raw)
	if strings.Contains(got, "token") {
		t.Fatalf("query should be stripped: %q", got)
	}
	if got != "https://api.example.com/v1/sign" {
		t.Fatalf("got %q", got)
	}
}

func TestSanitizeRequestURL_invalidTruncates(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", 250)
	got := SanitizeRequestURL("://" + long)
	if len(got) > 210 {
		t.Fatalf("expected truncation, len=%d", len(got))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("expected ellipsis suffix: %q", got)
	}
}

func TestDo_nonOK_truncatesBody(t *testing.T) {
	t.Parallel()
	large := strings.Repeat("Z", 5000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(large))
	}))
	t.Cleanup(srv.Close)

	client := srv.Client()
	_, err := Do(context.Background(), client, http.MethodGet, srv.URL+"/path?q=1", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if strings.Contains(msg, strings.Repeat("Z", 1000)) {
		t.Fatal("error should not contain a huge slice of the body")
	}
	if !strings.Contains(msg, "truncated") {
		t.Fatalf("expected truncation in message: %s", msg)
	}
	if strings.Contains(msg, "?q=1") {
		t.Fatalf("query should be stripped from URL in error: %s", msg)
	}
}
