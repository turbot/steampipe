package utils

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

func ParseCertificateInLocation(location string) (*x509.Certificate, error) {
	LogTime("db_local.parseCertificateInLocation start")
	defer LogTime("db_local.parseCertificateInLocation end")

	rootCertRaw, err := os.ReadFile(location)
	if err != nil {
		// if we can't read the certificate, then there's a problem with permissions
		return nil, err
	}
	// decode the pem blocks
	rootPemBlock, _ := pem.Decode(rootCertRaw)
	if rootPemBlock == nil {
		return nil, fmt.Errorf("could not decode PEM blocks from certificate at %s", location)
	}
	// parse the PEM Blocks to Certificates
	return x509.ParseCertificate(rootPemBlock.Bytes)
}

func WriteCertificate(path string, certificate []byte) error {
	return writeAsPEM(path, "CERTIFICATE", certificate)
}

func WritePrivateKey(path string, key *rsa.PrivateKey) error {
	return writeAsPEM(path, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
}

func writeAsPEM(location string, pemType string, b []byte) error {
	pemData := new(bytes.Buffer)
	err := pem.Encode(pemData, &pem.Block{
		Type:  pemType,
		Bytes: b,
	})
	if err != nil {
		log.Println("[INFO] Failed to encode to PEM")
		return err
	}
	if err := os.WriteFile(location, pemData.Bytes(), 0600); err != nil {
		log.Println("[INFO] Failed to save pem at", location)
		return err
	}
	return nil
}
