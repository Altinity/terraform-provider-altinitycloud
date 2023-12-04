package auth

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

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
