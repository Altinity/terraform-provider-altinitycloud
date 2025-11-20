package auth

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net/http"
	"time"

	sdkCrypto "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/crypto"
	sdkHttp "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/http"
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

func (a *Auth) GenerateCertificate(ctx context.Context, envName string) (string, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", err
	}
	csrPEM, err := createCertificateRequest(key, envName)
	if err != nil {
		return "", "", err
	}
	certPEM, err := a.signCertificateRequest(ctx, csrPEM)
	if err != nil {
		return "", "", err
	}
	keyPEM, err := sdkCrypto.EncodeRSAPrivateKey(key)
	if err != nil {
		return "", "", err
	}

	return string(certPEM), string(keyPEM), nil
}

func (a *Auth) signCertificateRequest(ctx context.Context, csrPEM []byte) ([]byte, error) {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("failed to get default HTTP transport")
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 defaultTransport.Proxy,
			DialContext:           defaultTransport.DialContext,
			ForceAttemptHTTP2:     defaultTransport.ForceAttemptHTTP2,
			MaxIdleConns:          defaultTransport.MaxIdleConns,
			IdleConnTimeout:       defaultTransport.IdleConnTimeout,
			TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
			ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
			TLSClientConfig: &tls.Config{
				RootCAs: a.RootCAs,
			},
		},
		Timeout: time.Minute,
	}
	url := fmt.Sprintf("%s/sign", a.URL)
	headers := map[string]string{}
	headers["Authorization"] = fmt.Sprintf("Bearer %s", a.APIToken)

	body, err := sdkHttp.Do(
		ctx,
		httpClient,
		http.MethodPost,
		url,
		headers,
		bytes.NewReader(csrPEM),
	)
	if err != nil {
		return nil, err
	}
	if _, err := sdkCrypto.DecodeCertificate(body); err != nil {
		return nil, fmt.Errorf("POST %s: parse body %q: %v", url, string(body), err)
	}
	return body, nil
}

func createCertificateRequest(pk crypto.PrivateKey, envName string) (csrPEM []byte, err error) {
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
	return sdkCrypto.EncodeCertificateRequestDER(csrDER)
}
