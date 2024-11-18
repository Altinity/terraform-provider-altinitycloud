package auth

import (
	"bytes"
	"encoding/pem"
)

func EncodeCertificateRequestDER(der []byte) ([]byte, error) {
	return Encode(der, "CERTIFICATE REQUEST")
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
