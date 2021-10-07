package db_local

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"path/filepath"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
)

func SslMode() string {
	certExists := helpers.FileExists(filepath.Join(getDataLocation(), constants.ServerCert))
	privateKeyExists := helpers.FileExists(filepath.Join(getDataLocation(), constants.ServerCertKey))
	if certExists && privateKeyExists {
		return "require"
	}
	return "disable"
}

func SslStatus() string {
	status := SslMode()
	if status == "require" {
		return "on"
	}
	return "off"
}

func writeCertFile(filePath string, cert string) error {
	return ioutil.WriteFile(filePath, []byte(cert), 0600)
}

func ensureSelfSignedCertificate() (err error) {
	// Check if the file exists if the file exists then do not generate the cert and key
	// Generate the certificate if there is no existing certificate
	certExists := helpers.FileExists(filepath.Join(getDataLocation(), constants.ServerCert))
	privateKeyExists := helpers.FileExists(filepath.Join(getDataLocation(), constants.ServerCertKey))

	if certExists && privateKeyExists {
		return nil
	}

	// Create your own certificate authority
	caPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println("[INFO] Private key creation failed for ca failed")
		return err
	}

	// Certificate authority input
	caInput := x509.Certificate{
		SerialNumber:          big.NewInt(2020),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(3, 0, 0),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &caInput, &caInput, &caPriv.PublicKey, caPriv)
	// err = fmt.Errorf("Failed")
	if err != nil {
		log.Println("[INFO] Failed to create certificate")
		return err
	}

	certPem := &bytes.Buffer{}

	pem.Encode(certPem, &pem.Block{Type: "CERTIFICATE", Bytes: cert})

	if err := writeCertFile(getRootCertLocation(), certPem.String()); err != nil {
		log.Println("[INFO] Failed to save the certificate")
		return err
	}

	privPem := new(bytes.Buffer)

	pem.Encode(privPem, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPriv)})

	// set up for server certificate
	serverCertInput := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName: string("localhost"),
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(3, 0, 0),
	}

	// Generate the server private key
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCertInput, &caInput, &serverPrivKey.PublicKey, caPriv)

	if err != nil {
		log.Println("[INFO] Failed to create server certificate")
		return err
	}

	// Encode and save the server certificate
	serverCertPem := new(bytes.Buffer)
	pem.Encode(serverCertPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	if err := writeCertFile(getServerCertLocation(), serverCertPem.String()); err != nil {
		log.Println("[INFO] Failed to save server certificate")
		return err
	}

	// Encode and save the server private key
	serverPrivKeyPem := new(bytes.Buffer)
	pem.Encode(serverPrivKeyPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	if err := writeCertFile(getServerCertKeyLocation(), serverPrivKeyPem.String()); err != nil {
		log.Println("[INFO] Failed to save server private key")
		return err
	}

	return nil
}
