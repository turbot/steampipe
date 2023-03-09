package db_common

import (
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/pkg/utils"
)

func AddRootCertToConfig(config *pgconn.Config, certLocation string) error {
	rootCert, err := utils.ParseCertificateInLocation(certLocation)
	if err != nil {
		return err
	}
	config.TLSConfig.RootCAs.AddCert(rootCert)
	return nil
}
