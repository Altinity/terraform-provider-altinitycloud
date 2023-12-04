package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Auth struct {
	RootCAs  *x509.CertPool
	URL      string
	APIToken string
}

func NewAuth(rootCAs *x509.CertPool, authUrl string, apiToken string) *Auth {
	return &Auth{
		RootCAs:  rootCAs,
		URL:      authUrl,
		APIToken: apiToken,
	}
}

func (a *Auth) GenerateCertificate(envName string) (string, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", err
	}
	csrPEM, err := createCertificateRequest(key, envName)
	if err != nil {
		return "", "", err
	}

	certPEM, err := do(context.Background(), &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: a.RootCAs,
			},
		},
		Timeout: time.Minute,
	}, fmt.Sprintf("%s/sign", a.URL), a.APIToken, csrPEM)
	if err != nil {
		return "", "", err
	}
	keyPEM, err := EncodeRSAPrivateKey(key)
	if err != nil {
		return "", "", err
	}

	return string(certPEM), string(keyPEM), nil
}

func createCertificateRequest(pk interface{}, envName string) (csrPEM []byte, err error) {
	req := x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA256WithRSA,
		Subject: pkix.Name{
			CommonName: envName,
		},
	}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &req, pk)
	if err != nil {
		return nil, err
	}
	return EncodeCertificateRequestDER(csrDER)
}

func do(ctx context.Context, httpClient *http.Client, url, apiToken string, csrPEM []byte) (certPEM []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		url, bytes.NewReader(csrPEM))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body) // keep-alive
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST %s resulted in %s %s", url, res.Status,
			string(body))
	}
	// Check response body is PEM-encoded.
	if _, err := DecodeCertificate(body); err != nil {
		return nil, fmt.Errorf("POST %s: parse body %q: %v", url, string(body), err)
	}
	return body, nil
}
