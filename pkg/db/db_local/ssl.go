package db_local

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/utils"
)

const CertIssuer = "steampipe.io"

var (
	CertExpiryTolerance      = 180 * (24 * time.Hour)     // 180 days
	RootCertValidityPeriod   = 5 * 365 * (24 * time.Hour) // 5 years
	ServerCertValidityPeriod = 365 * (24 * time.Hour)     // 1 year
)

// CertificatesExist checks if the root and server certificate and key files exist
func CertificatesExist() bool {
	return filehelpers.FileExists(getRootCertLocation()) && filehelpers.FileExists(getServerCertLocation())
}

// RemoveServerCertificate removes the server certificate certificates so it will be regenerated
func RemoveServerCertificate() error {
	utils.LogTime("db_local.RemoveServerCertificate start")
	defer utils.LogTime("db_local.RemoveServerCertificate end")

	if err := os.Remove(getServerCertLocation()); err != nil {
		return err
	}
	return os.Remove(getServerCertKeyLocation())
}

// RemoveAllCertificates removes root and server certificates so that they can be regenerated
func RemoveAllCertificates() error {
	utils.LogTime("db_local.RemoveAllCertificates start")
	defer utils.LogTime("db_local.RemoveAllCertificates end")

	// remove the root cert (but not key)
	if err := os.Remove(getRootCertLocation()); err != nil {
		return err
	}
	// remove the server cert and key
	return RemoveServerCertificate()
}

// ValidateRootCertificate checks the root certificate exists, is not expired and has correct Subject
func ValidateRootCertificate() bool {
	utils.LogTime("db_local.ValidateRootCertificates start")
	defer utils.LogTime("db_local.ValidateRootCertificates end")

	rootCertificate, err := utils.ParseCertificateInLocation(getRootCertLocation())
	if err != nil {
		return false
	}

	return (rootCertificate.Subject.CommonName == CertIssuer) && isCerticateExpiring(rootCertificate)
}

// ValidateServerCertificate checks the server certificate exists, is not expired and has correct issuer
func ValidateServerCertificate() bool {
	utils.LogTime("db_local.ValidateServerCertificates start")
	defer utils.LogTime("db_local.ValidateServerCertificates end")

	serverCertificate, err := utils.ParseCertificateInLocation(getServerCertLocation())
	if err != nil {
		return false
	}

	return (serverCertificate.Issuer.CommonName == CertIssuer) && isCerticateExpiring(serverCertificate)
}

// if certificate or private key files do not exist, generate them
func ensureSelfSignedCertificate() (err error) {
	if serverCertificateAndKeyExist() && rootCertificateAndKeyExists() {
		return nil
	}

	// so one or both of the root and server certificate need creating
	var rootPrivateKey *rsa.PrivateKey
	var rootCertificate *x509.Certificate
	if rootCertificateAndKeyExists() {
		// if the root cert and key exist, load them
		rootPrivateKey, err = loadRootPrivateKey()
		if err != nil {
			return err
		}
		rootCertificate, err = utils.ParseCertificateInLocation(getRootCertLocation())
	} else {
		// otherwise generate them
		rootCertificate, rootPrivateKey, err = generateRootCertificate()
	}
	if err != nil {
		return err
	}

	// now generate new server cert
	return generateServerCertificates(rootCertificate, rootPrivateKey)

}

// rootCertificateAndKeyExists checks if the root certificate ands private key files exist
func rootCertificateAndKeyExists() bool {
	return filehelpers.FileExists(getRootCertLocation()) && filehelpers.FileExists(getRootCertKeyLocation())
}

// serverCertificateAndKeyExist checks if the server certificate ands private key files exist
func serverCertificateAndKeyExist() bool {
	return filehelpers.FileExists(getServerCertLocation()) && filehelpers.FileExists(getServerCertKeyLocation())
}

// isCerticateExpiring checks whether the certificate expires within a predefined CertExpiryTolerance period (defined above)
func isCerticateExpiring(certificate *x509.Certificate) bool {
	return certificate.NotAfter.Add(-CertExpiryTolerance).After(time.Now())
}

func generateRootCertificate() (*x509.Certificate, *rsa.PrivateKey, error) {
	utils.LogTime("db_local.generateServiceCertificates start")
	defer utils.LogTime("db_local.generateServiceCertificates end")

	// Load or create our own certificate authority
	caPrivateKey, err := ensureRootPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()

	// Certificate authority input
	caCertificateData := &x509.Certificate{
		SerialNumber:          big.NewInt(2020),
		NotBefore:             now,
		NotAfter:              now.Add(RootCertValidityPeriod),
		Subject:               pkix.Name{CommonName: CertIssuer},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caCertificate, err := x509.CreateCertificate(rand.Reader, caCertificateData, caCertificateData, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Println("[WARN] failed to create certificate")
		return nil, nil, err
	}

	if err := utils.WriteCertificate(getRootCertLocation(), caCertificate); err != nil {
		log.Println("[WARN] failed to save the certificate")
		return nil, nil, err
	}

	return caCertificateData, caPrivateKey, nil
}

func generateServerCertificates(caCertificateData *x509.Certificate, caPrivateKey *rsa.PrivateKey) error {
	utils.LogTime("db_local.generateServerCertificates start")
	defer utils.LogTime("db_local.generateServerCertificates end")

	now := time.Now()

	// set up for server certificate
	serverCertificateData := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject:      caCertificateData.Subject,
		Issuer:       caCertificateData.Subject,
		NotBefore:    now,
		NotAfter:     now.Add(ServerCertValidityPeriod),
	}

	// Generate the server private key
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCertificateData, caCertificateData, &serverPrivKey.PublicKey, caPrivateKey)

	if err != nil {
		log.Println("[INFO] Failed to create server certificate")
		return err
	}

	if err := utils.WriteCertificate(getServerCertLocation(), serverCertBytes); err != nil {
		log.Println("[INFO] Failed to save server certificate")
		return err
	}
	if err := utils.WritePrivateKey(getServerCertKeyLocation(), serverPrivKey); err != nil {
		log.Println("[INFO] Failed to save server private key")
		return err
	}

	return nil
}

// derive ssl status from out ssl mode
func sslStatus() string {
	if serverCertificateAndKeyExist() {
		return "on"
	}
	return "off"
}

// derive ssl parameters from the presence of the server certificate and key file
func dsnSSLParams() map[string]string {
	if serverCertificateAndKeyExist() && rootCertificateAndKeyExists() {
		// as per https://www.postgresql.org/docs/current/libpq-ssl.html#LIBQ-SSL-CERTIFICATES :
		//
		// For backwards compatibility with earlier versions of PostgreSQL, if a root CA file exists, the
		// behavior of sslmode=require will be the same as that of verify-ca, meaning the
		// server certificate is validated against the CA. Relying on this behavior is discouraged, and
		// applications that need certificate validation should always use verify-ca or verify-full.
		//
		// Since we are using the Root Certificate, 'require' is overridden with 'verify-ca' anyway
		return map[string]string{
			"sslmode":     "verify-ca",
			"sslrootcert": getRootCertLocation(),
			"sslcert":     getServerCertLocation(),
			"sslkey":      getServerCertKeyLocation(),
		}
	}
	return map[string]string{"sslmode": "disable"}
}

func ensureRootPrivateKey() (*rsa.PrivateKey, error) {
	// first try to load the key
	// if any errors are encountered this will just return nil
	caPrivateKey, _ := loadRootPrivateKey()
	if caPrivateKey != nil {
		// we loaded one
		return caPrivateKey, nil
	}
	// so we failed to load the key - generate instead
	var err error
	caPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println("[WARN] private key creation failed for ca failed")
		return nil, err
	}
	if err := utils.WritePrivateKey(getRootCertKeyLocation(), caPrivateKey); err != nil {
		log.Println("[WARN] failed to save root private key")
		return nil, err
	}
	return caPrivateKey, nil
}

func loadRootPrivateKey() (*rsa.PrivateKey, error) {
	location := getRootCertKeyLocation()

	priv, err := os.ReadFile(location)
	if err != nil {
		log.Printf("[TRACE] loadRootPrivateKey - failed to load key from %s: %s", location, err.Error())
		return nil, err
	}

	privPem, _ := pem.Decode(priv)
	if privPem.Type != "RSA PRIVATE KEY" {
		log.Printf("[TRACE] RSA private key is of the wrong type: %v", privPem.Type)
		return nil, fmt.Errorf("RSA private key is of the wrong type: %v", privPem.Type)
	}

	privPemBytes := privPem.Bytes

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil {
			// note this returns type `interface{}`
			log.Printf("[TRACE] failed to parse RSA private key: %s", err.Error())
			return nil, err
		}
	}

	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		log.Printf("[TRACE] failed to parse RSA private key")
		return nil, fmt.Errorf("failed to parse RSA private key")
	}
	return privateKey, nil
}
