package task

import (
	"log"

	"github.com/turbot/steampipe/db/db_local"
)

func validateServiceCertificates() {
	if db_local.CertificatesExist() {
		// don't do anything - certificates haven't been installed yet
		return
	}

	if !db_local.ValidateServiceCertificates() {
		err := db_local.RemoveServiceCertificates()
		if err != nil {
			log.Println("[TRACE] there was an issue removing invalid certificates in TaskRunner", err)
		}
	}
}
