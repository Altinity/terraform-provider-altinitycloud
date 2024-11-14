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
	"net/http"
	"time"

	crypto "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/crypto"
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

func (a *Auth) GenerateCertificate(envName string) (string, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", err
	}
	csrPEM, err := a.createCertificateRequest(key, envName)
	if err != nil {
		return "", "", err
	}
	certPEM, err := a.signCertificateRequest(context.Background(), csrPEM)
	if err != nil {
		return "", "", err
	}
	keyPEM, err := crypto.EncodeRSAPrivateKey(key)
	if err != nil {
		return "", "", err
	}

	return string(certPEM), string(keyPEM), nil
}

func (a *Auth) createCertificateRequest(pk interface{}, envName string) (csrPEM []byte, err error) {
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
	return crypto.EncodeCertificateRequestDER(csrDER)
}

func (a *Auth) signCertificateRequest(ctx context.Context, csrPEM []byte) ([]byte, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
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
	if _, err := crypto.DecodeCertificate(body); err != nil {
		return nil, fmt.Errorf("POST %s: parse body %q: %v", url, string(body), err)
	}
	return body, nil
}
