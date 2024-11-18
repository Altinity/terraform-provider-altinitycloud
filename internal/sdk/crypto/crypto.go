package crypto

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"

	sdkHttp "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/http"
)

type Crypto struct {
	RootCAs *x509.CertPool
	URL     string
}

func NewCrypto(rootCAs *x509.CertPool, cryptoUrl string) *Crypto {
	return &Crypto{
		RootCAs: rootCAs,
		URL:     cryptoUrl,
	}
}

func (c *Crypto) Encrypt(pem string, value string) (string, error) {
	split, err := split([]byte(pem))
	if err != nil {
		return "", err
	}
	if len(split) != 2 {
		return "", fmt.Errorf("malformed %s: expected 2 PEMs, instead got %d", pem, len(split))
	}
	tlsCert, err := x509KeyPairWithLeaf(split[0], split[1])
	if err != nil {
		return "", err
	}

	res, err := c.fetchPublicKey(context.Background(), tlsCert)
	if err != nil {
		return string(split[0]) + string(split[1]), err
	}
	key, err := parseRSAPublicKey(res)
	if err != nil {
		return "", err
	}
	v, err := encryptWithRSAPublicKey(value, key)
	if err != nil {
		return "", err
	}

	return v, nil
}

func (c *Crypto) fetchPublicKey(ctx context.Context, tlsCert tls.Certificate) (pem []byte, err error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: c.RootCAs,
				Certificates: []tls.Certificate{
					tlsCert,
				},
			},
		},
		Timeout: time.Minute,
	}
	url := fmt.Sprintf("%s/key", c.URL)
	body, err := sdkHttp.Do(ctx, httpClient, http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, err
	}
	if _, err := decodeRSAPublicKey(body); err != nil {
		return nil, fmt.Errorf("GET %s: parse body %q: %v", url, string(body), err)
	}
	return body, nil
}

func split(data []byte) ([][]byte, error) {
	var blocks [][]byte
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		encodedBlock := pem.EncodeToMemory(block)
		blocks = append(blocks, encodedBlock)
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no PEM blocks found")
	}

	return blocks, nil
}

func x509KeyPairWithLeaf(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
	tlsCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return tls.Certificate{}, err
	}
	leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return tls.Certificate{}, err
	}
	tlsCert.Leaf = leaf
	return tlsCert, nil
}

func decodeRSAPublicKey(data []byte) (*pem.Block, error) {
	return Decode(data, "RSA PUBLIC KEY")
}

func parseRSAPublicKey(data []byte) (*rsa.PublicKey, error) {
	block, err := Decode(data, "RSA PUBLIC KEY")
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

func encryptWithRSAPublicKey(token string, pub *rsa.PublicKey) (string, error) {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, []byte(token), nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%x", fingerprint(pub), ciphertext), nil
}

func fingerprint(pub *rsa.PublicKey) string {
	h := md5.New()
	h.Write(x509.MarshalPKCS1PublicKey(pub))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func DecodeCertificate(data []byte) (*pem.Block, error) {
	return Decode(data, "CERTIFICATE")
}

func LoadCertPool(cert string) (*x509.CertPool, error) {
	clientCAs := x509.NewCertPool()
	ok := clientCAs.AppendCertsFromPEM([]byte(cert))
	if !ok {
		return nil, errors.New("pem: failed to append certificates")
	}
	return clientCAs, nil
}

func Decode(data []byte, blockType string) (*pem.Block, error) {
	p, _ := pem.Decode(data)
	if p == nil {
		return nil, errors.New("pem: invalid")
	}
	if p.Type != blockType {
		return nil, fmt.Errorf("pem: expected type %s, got %s", blockType, p.Type)
	}
	return p, nil
}

func EncodeCertificateRequestDER(der []byte) ([]byte, error) {
	return Encode(der, "CERTIFICATE REQUEST")
}

func EncodeRSAPrivateKey(p *rsa.PrivateKey) ([]byte, error) {
	key := x509.MarshalPKCS1PrivateKey(p)
	return Encode(key, "RSA PRIVATE KEY")
}

func Encode(data []byte, blockType string) ([]byte, error) {
	var out bytes.Buffer
	if err := pem.Encode(&out, &pem.Block{
		Type:  blockType,
		Bytes: data,
	}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
