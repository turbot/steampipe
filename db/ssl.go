package db

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
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/constants"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		log.Println("File exists\n")
		return true
	}
	log.Println("File does not exist\n")
	return false

}

func SslMode() string {
	certExists := fileExists(filepath.Join(getDataLocation(), constants.ServerCert))
	privateKeyExists := fileExists(filepath.Join(getDataLocation(), constants.ServerKey))
	if certExists && privateKeyExists {
		return "require"
	}
	return "disable"
}

func SslStatus() string {
	certExists := fileExists(filepath.Join(getDataLocation(), constants.ServerCert))
	privateKeyExists := fileExists(filepath.Join(getDataLocation(), constants.ServerKey))
	if certExists && privateKeyExists {
		return "on"
	}
	return "off"
}

func writeCertFile(filePath string, cert string) error {
	return ioutil.WriteFile(filePath, []byte(cert), 0600)
}

func generateSelfSignedCertificate() (err error) {

	// Check if the file exists if the file exists then do not generate the cert and key
	certExists := fileExists(filepath.Join(getDataLocation(), constants.ServerCert))
	privateKeyExists := fileExists(filepath.Join(getDataLocation(), constants.ServerKey))

	if certExists && privateKeyExists {
		return nil
	}

	// Create your own certificate authority
	caPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println("Private key creation failed")
		return err
	}

	caInput := x509.Certificate{
		SerialNumber: big.NewInt(2020),

		Subject: pkix.Name{
			Organization: []string{"Turbot"},
			Country:      []string{"US"},
			Province:     []string{""},
			Locality:     []string{"New Jersey"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &caInput, &caInput, &caPriv.PublicKey, caPriv)

	if err != nil {
		log.Println("Failed to create certificate")
		return err
	}

	certPem := &bytes.Buffer{}

	pem.Encode(certPem, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	err = writeCertFile(filepath.Join(getDataLocation(), constants.RootCert), certPem.String())

	if err != nil {
		log.Println("Failed to write certificate")
		return err
	}

	privPem := new(bytes.Buffer)

	pem.Encode(privPem, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPriv)})

	if err != nil {
		log.Println("Failed to write private key")
		return err
	}

	// set up our server certificate
	serverCertInput := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName:   string("localhost"),
			Organization: []string{"Steampipe"},
			Country:      []string{"US"},
			Province:     []string{""},
			Locality:     []string{"New Jersey"},
		},
		// IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
		// SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// Generate the server private key
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCertInput, &caInput, &serverPrivKey.PublicKey, caPriv)

	if err != nil {
		log.Println("Failed to create server certificate")
		return err
	}

	// Encode and save the server certificate
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	err = writeCertFile(filepath.Join(getDataLocation(), constants.ServerCert), certPEM.String())

	if err != nil {
		log.Println("Failed to write server crt", err)
		return err
	}

	// Encode and save the server private key
	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})
	err = writeCertFile(filepath.Join(getDataLocation(), constants.ServerKey), certPrivKeyPEM.String())

	if err != nil {
		log.Println("Failed to write server private key", err)
		return err
	}
	return nil
}
