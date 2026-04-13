package http

import (
	"fmt"
	"net/url"
	"strings"
)

// maxErrorBodyPreview limits how much of a response body is included in errors
// to avoid dumping large HTML/error pages into Terraform CLI output.
const maxErrorBodyPreview = 512

// PreviewBodyForError returns a short, single-line representation of body for use
// in error messages. Long bodies are truncated with a total byte count.
func PreviewBodyForError(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	n := len(body)
	chunk := body
	if n > maxErrorBodyPreview {
		chunk = body[:maxErrorBodyPreview]
	}
	s := strings.ReplaceAll(string(chunk), "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if n > maxErrorBodyPreview {
		return fmt.Sprintf("%s… (truncated, %d bytes total)", s, n)
	}
	return s
}

// SanitizeRequestURL strips query and fragment from a URL string so secrets in
// query parameters are less likely to appear in error messages. If parsing
// fails, the string is truncated to a safe length.
func SanitizeRequestURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		const maxInvalidURLLen = 200
		if len(raw) > maxInvalidURLLen {
			return raw[:maxInvalidURLLen] + "…"
		}
		return raw
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
