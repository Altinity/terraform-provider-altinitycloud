package auth

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func EncodeCertificateRequestDER(der []byte) ([]byte, error) {
	return Encode(der, "CERTIFICATE REQUEST")
}

func EncodeRSAPrivateKey(p *rsa.PrivateKey) ([]byte, error) {
	return Encode(x509.MarshalPKCS1PrivateKey(p), "RSA PRIVATE KEY")
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
