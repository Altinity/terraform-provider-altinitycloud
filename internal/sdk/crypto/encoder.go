package crypto

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

func EncodeCertificateRequestDER(der []byte) ([]byte, error) {
	return encode(der, "CERTIFICATE REQUEST")
}

func DecodeCertificate(data []byte) (*pem.Block, error) {
	return decode(data, "CERTIFICATE")
}

func EncodeRSAPrivateKey(p *rsa.PrivateKey) ([]byte, error) {
	key := x509.MarshalPKCS1PrivateKey(p)
	return encode(key, "RSA PRIVATE KEY")
}

func DecodeRSAPublicKey(data []byte) (*pem.Block, error) {
	return decode(data, "RSA PUBLIC KEY")
}

func ParseRSAPublicKey(data []byte) (*rsa.PublicKey, error) {
	block, err := DecodeRSAPublicKey(data)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

func encode(data []byte, blockType string) ([]byte, error) {
	var out bytes.Buffer
	if err := pem.Encode(&out, &pem.Block{
		Type:  blockType,
		Bytes: data,
	}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func decode(data []byte, blockType string) (*pem.Block, error) {
	p, _ := pem.Decode(data)
	if p == nil {
		return nil, errors.New("pem: invalid")
	}
	if p.Type != blockType {
		return nil, fmt.Errorf("pem: expected type %s, got %s", blockType, p.Type)
	}
	return p, nil
}
