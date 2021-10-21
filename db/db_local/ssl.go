package db_local

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/utils"
)

const CertIssuer = "steampipe.io"

var (
	CertExpiryTolerance      = 180 * (24 * time.Hour) // 180 days
	RootCertValidityPeriod   = 365 * (24 * time.Hour) // 1 year
	ServerCertValidityPeriod = 365 * (24 * time.Hour) // 1 year
)

func ensureSelfSignedCertificate() (err error) {
	// Check if the file exists if the file exists then do not generate the cert and key
	// Generate the certificate if there is no existing certificate
	certExists := helpers.FileExists(getServerCertLocation())
	privateKeyExists := helpers.FileExists(getServerCertKeyLocation())

	if certExists && privateKeyExists {
		return nil
	}

	return generateServiceCertificates()
}

func sslStatus() string {
	status := sslMode()
	if status == "require" {
		return "on"
	}
	return "off"
}

func sslMode() string {
	certExists := helpers.FileExists(getServerCertLocation())
	privateKeyExists := helpers.FileExists(getServerCertKeyLocation())
	if certExists && privateKeyExists {
		return "require"
	}
	return "disable"
}

// CertificatesExist checks if the root and server certificate files exist
func CertificatesExist() bool {
	return helpers.FileExists(getRootCertLocation()) && helpers.FileExists(getServerCertLocation())
}

// RemoveServiceCertificates removes generated certificates so that they can be regenerated
func RemoveServiceCertificates() error {
	utils.LogTime("db_local.RemoveServiceCertificates start")
	defer utils.LogTime("db_local.RemoveServiceCertificates end")

	if err := os.Remove(getServerCertLocation()); err != nil {
		return err
	}
	if err := os.Remove(getServerCertKeyLocation()); err != nil {
		return err
	}
	if err := os.Remove(getRootCertLocation()); err != nil {
		return err
	}
	return nil
}

// ValidateServiceCertificates checks that both the root and server certificates satisfy the following conditions:
// * rootCertificate Subject CN is equal to CertIssuer (defined above)
// * serverCertificate Issuer CN is equal to CertIssuer (defined above)
// * both server and root certificates haven't expired or are about to expire
func ValidateServiceCertificates() bool {
	utils.LogTime("db_local.ValidateServiceCertificates start")
	defer utils.LogTime("db_local.ValidateServiceCertificates end")

	rootCertificate, err := parseCertificateInLocation(getRootCertLocation())
	if err != nil {
		return false
	}
	serverCertificate, err := parseCertificateInLocation(getServerCertLocation())
	if err != nil {
		return false
	}

	return ((rootCertificate.Subject.CommonName == CertIssuer) &&
		(serverCertificate.Issuer.CommonName == CertIssuer) &&
		(isCerticateExpiring(rootCertificate) ||
			isCerticateExpiring(serverCertificate)))
}

// isCerticateExpiring checks whether the certificate expires within a predefined CertExpiryTolerance period (defined above)
func isCerticateExpiring(certificate *x509.Certificate) bool {
	return certificate.NotAfter.Add(CertExpiryTolerance).After(time.Now())
}

func parseCertificateInLocation(location string) (*x509.Certificate, error) {
	utils.LogTime("db_local.parseCertificateInLocation start")
	defer utils.LogTime("db_local.parseCertificateInLocation end")

	rootCertRaw, err := ioutil.ReadFile(getRootCertLocation())
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

func generateServiceCertificates() error {
	utils.LogTime("db_local.generateServiceCertificates start")
	defer utils.LogTime("db_local.generateServiceCertificates end")

	// Create our own certificate authority
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println("[INFO] Private key creation failed for ca failed")
		return err
	}

	NOW := time.Now()

	// Certificate authority input
	caCertificateData := x509.Certificate{
		SerialNumber:          big.NewInt(2020),
		NotBefore:             NOW,
		NotAfter:              NOW.Add(ServerCertValidityPeriod),
		Subject:               pkix.Name{CommonName: CertIssuer},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caCertificate, err := x509.CreateCertificate(rand.Reader, &caCertificateData, &caCertificateData, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Println("[INFO] Failed to create certificate")
		return err
	}

	caCertificatePem := &bytes.Buffer{}
	err = pem.Encode(caCertificatePem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertificate,
	})
	if err != nil {
		log.Println("[INFO] Failed to encode to PEM")
		return err
	}

	if err := writeCertFile(getRootCertLocation(), caCertificatePem.String()); err != nil {
		log.Println("[INFO] Failed to save the certificate")
		return err
	}

	caPrivateKeyPem := new(bytes.Buffer)
	err = pem.Encode(caPrivateKeyPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})
	if err != nil {
		log.Println("[INFO] Failed to encode to PEM for ca private key")
		return err
	}
	if err := writeCertFile(getRootCertKeyLocation(), caPrivateKeyPem.String()); err != nil {
		log.Println("[INFO] Failed to save root private key")
		return err
	}

	// set up for server certificate
	serverCertificateData := x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject:      caCertificateData.Subject,
		Issuer:       caCertificateData.Subject,
		NotBefore:    NOW,
		NotAfter:     NOW.Add(ServerCertValidityPeriod),
	}

	// Generate the server private key
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &serverCertificateData, &caCertificateData, &serverPrivKey.PublicKey, caPrivateKey)

	if err != nil {
		log.Println("[INFO] Failed to create server certificate")
		return err
	}

	// Encode and save the server certificate
	serverCertPem := new(bytes.Buffer)
	err = pem.Encode(serverCertPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	if err != nil {
		log.Println("[INFO] Failed to encode to PEM for server certificate")
		return err
	}
	if err := writeCertFile(getServerCertLocation(), serverCertPem.String()); err != nil {
		log.Println("[INFO] Failed to save server certificate")
		return err
	}

	// Encode and save the server private key
	serverPrivKeyPem := new(bytes.Buffer)
	err = pem.Encode(serverPrivKeyPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})
	if err != nil {
		log.Println("[INFO] Failed to encode to PEM for server private key")
		return err
	}
	if err := writeCertFile(getServerCertKeyLocation(), serverPrivKeyPem.String()); err != nil {
		log.Println("[INFO] Failed to save server private key")
		return err
	}

	return nil
}

func writeCertFile(filePath string, cert string) error {
	return ioutil.WriteFile(filePath, []byte(cert), 0600)
}
