package crypto

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

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
		return "", fmt.Errorf("malformed PEM: expected 2 blocks (certificate + key), instead got %d", len(split))
	}
	tlsCert, err := x509KeyPairWithLeaf(split[0], split[1])
	if err != nil {
		return "", err
	}
	res, err := c.fetchPublicKey(context.Background(), tlsCert)
	if err != nil {
		return "", err
	}
	key, err := ParseRSAPublicKey(res)
	if err != nil {
		return "", err
	}
	v, err := encryptWithRSAPublicKey(value, key)
	if err != nil {
		return "", err
	}

	return v, nil
}

func (c *Crypto) Decrypt(pkPem string, value string) (string, error) {
	block, _ := pem.Decode([]byte(pkPem))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block from private key")
	}
	pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed encrypted value: expected format \"fingerprint.ciphertext\"")
	}
	if parts[0] != fingerprint(&pk.PublicKey) {
		return "", fmt.Errorf("token encrypted with unknown key: %s", parts[0])
	}
	hash := sha256.New()
	enc, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	cleartext, err := rsa.DecryptOAEP(hash, rand.Reader, pk, enc, nil)
	if err != nil {
		return "", err
	}
	return string(cleartext), nil
}

func (c *Crypto) fetchPublicKey(ctx context.Context, tlsCert tls.Certificate) (pem []byte, err error) {
	httpClient, err := sdkHttp.NewClient(c.RootCAs, tlsCert)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/key", c.URL)
	body, err := sdkHttp.Do(ctx, httpClient, http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, err
	}
	if _, err := DecodeRSAPublicKey(body); err != nil {
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
