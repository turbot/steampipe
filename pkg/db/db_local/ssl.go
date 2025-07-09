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
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/db/sslio"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

const (
	CertIssuer               = "steampipe.io"
	ServerCertValidityPeriod = 3 * (365 * (24 * time.Hour)) // 3 years
)

var EndOfTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

func removeExpiringSelfIssuedCertificates() error {
	if !certificatesExist() {
		// don't do anything - certificates haven't been installed yet
		return nil
	}

	if isRootCertificateExpiring() && !isRootCertificateSelfIssued() {
		return sperr.New("cannot rotate certificate not issue by steampipe")
	}

	if isServerCertificateExpiring() && !isServerCertificateSelfIssued() {
		return sperr.New("cannot rotate certificate not issue by steampipe")
	}

	if isRootCertificateExpiring() {
		// if root certificate is not valid (i.e. expired), remove root and server certs,
		// they will both be regenerated
		err := removeAllCertificates()
		if err != nil {
			return sperr.WrapWithRootMessage(err, "issue removing invalid root certificate")
		}
	} else if isServerCertificateExpiring() {
		// if server certificate is not valid (i.e. expired), remove it,
		// it will be regenerated
		err := removeServerCertificate()
		if err != nil {
			return sperr.WrapWithRootMessage(err, "issue removing invalid server certificate")
		}
	}
	return nil
}

func isRootCertificateSelfIssued() bool {
	rootCertificate, err := sslio.ParseCertificateInLocation(filepaths.GetRootCertLocation())
	if err != nil {
		return false
	}
	return rootCertificate.IsCA && strings.EqualFold(rootCertificate.Subject.CommonName, CertIssuer)
}

func isServerCertificateSelfIssued() bool {
	serverCertificate, err := sslio.ParseCertificateInLocation(filepaths.GetServerCertLocation())
	if err != nil {
		return false
	}
	return !serverCertificate.IsCA && strings.EqualFold(serverCertificate.Issuer.CommonName, CertIssuer)
}

// certificatesExist checks if the root and server certificate and key files exist
func certificatesExist() bool {
	return filehelpers.FileExists(filepaths.GetRootCertLocation()) && filehelpers.FileExists(filepaths.GetServerCertLocation())
}

// removeServerCertificate removes the server certificate certificates so it will be regenerated
func removeServerCertificate() error {
	utils.LogTime("db_local.RemoveServerCertificate start")
	defer utils.LogTime("db_local.RemoveServerCertificate end")

	if err := os.Remove(filepaths.GetServerCertLocation()); err != nil {
		return err
	}
	return os.Remove(filepaths.GetServerCertKeyLocation())
}

// removeAllCertificates removes root and server certificates so that they can be regenerated
func removeAllCertificates() error {
	utils.LogTime("db_local.RemoveAllCertificates start")
	defer utils.LogTime("db_local.RemoveAllCertificates end")

	// remove the root cert (but not key)
	if err := os.Remove(filepaths.GetRootCertLocation()); err != nil {
		return err
	}
	// remove the server cert and key
	return removeServerCertificate()
}

// isRootCertificateExpiring checks the root certificate exists, is not expired and has correct Subject
func isRootCertificateExpiring() bool {
	utils.LogTime("db_local.isRootCertificateExpiring start")
	defer utils.LogTime("db_local.isRootCertificateExpiring end")
	rootCertificate, err := sslio.ParseCertificateInLocation(filepaths.GetRootCertLocation())
	if err != nil {
		return false
	}
	return isCerticateExpiring(rootCertificate)
}

// isServerCertificateExpiring checks the server certificate exists, is not expired and has correct issuer
func isServerCertificateExpiring() bool {
	utils.LogTime("db_local.ValidateServerCertificates start")
	defer utils.LogTime("db_local.ValidateServerCertificates end")
	serverCertificate, err := sslio.ParseCertificateInLocation(filepaths.GetServerCertLocation())
	if err != nil {
		return false
	}
	expiring := isCerticateExpiring(serverCertificate)
	return expiring
}

// if certificate or private key files do not exist, generate them
func ensureCertificates() (err error) {
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
		rootCertificate, err = sslio.ParseCertificateInLocation(filepaths.GetRootCertLocation())
	} else {
		// otherwise generate them
		rootCertificate, rootPrivateKey, err = generateRootCertificate()
	}
	if err != nil {
		return err
	}

	// now generate new server cert
	return generateServerCertificate(rootCertificate, rootPrivateKey)
}

// rootCertificateAndKeyExists checks if the root certificate ands private key files exist
func rootCertificateAndKeyExists() bool {
	return filehelpers.FileExists(filepaths.GetRootCertLocation()) && filehelpers.FileExists(filepaths.GetRootCertKeyLocation())
}

// serverCertificateAndKeyExist checks if the server certificate ands private key files exist
func serverCertificateAndKeyExist() bool {
	return filehelpers.FileExists(filepaths.GetServerCertLocation()) && filehelpers.FileExists(filepaths.GetServerCertKeyLocation())
}

// isCerticateExpiring checks whether the certificate expires within a predefined CertExpiryTolerance period (defined above)
func isCerticateExpiring(certificate *x509.Certificate) bool {
	// has the certificate elapsed 3/4 of its lifetime
	notBefore := certificate.NotBefore
	notAfter := certificate.NotAfter
	maxAllowedAge := float64(notAfter.Sub(notBefore)) * (0.75)
	currentAge := float64(time.Since(notBefore))

	// has current age exceeded the maximum allowed age
	return currentAge > maxAllowedAge
}

// generateRootCertificate generates a CA certificate along with a Private key
// the CA certificate sign itself
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
		SerialNumber:          getSerialNumber(now),
		NotBefore:             now,
		NotAfter:              EndOfTime,
		Subject:               pkix.Name{CommonName: CertIssuer},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caCertificate, err := x509.CreateCertificate(rand.Reader, caCertificateData, caCertificateData, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Println("[WARN] failed to create certificate")
		return nil, nil, err
	}

	if err := sslio.WriteCertificate(filepaths.GetRootCertLocation(), caCertificate); err != nil {
		log.Println("[WARN] failed to save the certificate")
		return nil, nil, err
	}

	return caCertificateData, caPrivateKey, nil
}

// generateServerCertificate creates a certificate signed by the CA certificate
func generateServerCertificate(caCertificateData *x509.Certificate, caPrivateKey *rsa.PrivateKey) error {
	utils.LogTime("db_local.generateServerCertificates start")
	defer utils.LogTime("db_local.generateServerCertificates end")

	now := time.Now()

	// set up for server certificate
	serverCertificateData := &x509.Certificate{
		SerialNumber: getSerialNumber(now),
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

	if err := sslio.WriteCertificate(filepaths.GetServerCertLocation(), serverCertBytes); err != nil {
		log.Println("[INFO] Failed to save server certificate")
		return err
	}
	if err := sslio.WritePrivateKey(filepaths.GetServerCertKeyLocation(), serverPrivKey); err != nil {
		log.Println("[INFO] Failed to save server private key")
		return err
	}

	return nil
}

// getSerialNumber generates a serial number for the certificate based on the passed in time in the format YYYYMMDD
func getSerialNumber(t time.Time) *big.Int {
	serialNumber, _ := strconv.ParseInt(
		t.Format("20060102"),
		10,
		64,
	)
	return big.NewInt(serialNumber)
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

		dsnSSLParams := map[string]string{
			"sslmode":     "verify-ca",
			"sslrootcert": filepaths.GetRootCertLocation(),
			"sslcert":     filepaths.GetServerCertLocation(),
			"sslkey":      filepaths.GetServerCertKeyLocation(),
		}

		if sslpassword := viper.GetString(constants.ArgDatabaseSSLPassword); sslpassword != "" {
			dsnSSLParams["sslpassword"] = sslpassword
		}

		return dsnSSLParams
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
	if err := sslio.WritePrivateKey(filepaths.GetRootCertKeyLocation(), caPrivateKey); err != nil {
		log.Println("[WARN] failed to save root private key")
		return nil, err
	}
	return caPrivateKey, nil
}

func loadRootPrivateKey() (*rsa.PrivateKey, error) {
	location := filepaths.GetRootCertKeyLocation()

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
