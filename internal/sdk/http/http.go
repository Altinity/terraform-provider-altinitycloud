package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NewClient creates an *http.Client that clones http.DefaultTransport settings
// and applies the given TLS configuration. This avoids duplicating transport
// setup across provider, auth, and crypto packages.
func NewClient(rootCAs *x509.CertPool, certs ...tls.Certificate) (*http.Client, error) {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("failed to get default HTTP transport")
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 defaultTransport.Proxy,
			DialContext:           defaultTransport.DialContext,
			ForceAttemptHTTP2:     defaultTransport.ForceAttemptHTTP2,
			MaxIdleConns:          defaultTransport.MaxIdleConns,
			IdleConnTimeout:       defaultTransport.IdleConnTimeout,
			TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
			ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
			TLSClientConfig: &tls.Config{
				RootCAs:      rootCAs,
				Certificates: certs,
			},
		},
		Timeout: time.Minute,
	}, nil
}

func Do(
	ctx context.Context,
	httpClient *http.Client,
	method, url string,
	headers map[string]string,
	requestBody io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		safeURL := SanitizeRequestURL(url)
		preview := PreviewBodyForError(body)
		if preview == "" {
			return nil, fmt.Errorf("%s %s resulted in %s", method, safeURL, res.Status)
		}
		return nil, fmt.Errorf("%s %s resulted in %s: %s", method, safeURL, res.Status, preview)
	}
	return body, nil
}
