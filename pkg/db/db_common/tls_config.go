package db_common

import (
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/v2/pkg/db/sslio"
)

func AddRootCertToConfig(config *pgconn.Config, certLocation string) error {
	rootCert, err := sslio.ParseCertificateInLocation(certLocation)
	if err != nil {
		return err
	}
	config.TLSConfig.RootCAs.AddCert(rootCert)
	return nil
}
