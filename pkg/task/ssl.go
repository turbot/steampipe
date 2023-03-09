package task

import (
	"log"

	"github.com/turbot/steampipe/pkg/db/db_local"
)

func validateServiceCertificates() {
	if !db_local.CertificatesExist() {
		// don't do anything - certificates haven't been installed yet
		return
	}

	if !db_local.ValidateRootCertificate() {
		// if root certificate is not valid (i.e. expired), remove root and server certs,
		// they will both be regenerated
		err := db_local.RemoveAllCertificates()
		if err != nil {
			log.Println("[TRACE] there was an issue removing invalid certificates in TaskRunner", err)
		}
	} else if !db_local.ValidateServerCertificate() {
		// if server certificate is not valid (i.e. expired), remove it,
		// it will be regenerated
		err := db_local.RemoveServerCertificate()
		if err != nil {
			log.Println("[TRACE] there was an issue removing invalid certificates in TaskRunner", err)
		}
	}
}
