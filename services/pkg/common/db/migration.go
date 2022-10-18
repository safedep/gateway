package db

import (
	"log"

	"github.com/safedep/gateway/services/pkg/common/db/adapters"
	"github.com/safedep/gateway/services/pkg/common/db/models"
)

func MigrateSqlModels(adapter adapters.SqlDataAdapter) error {
	db, err := adapter.GetDB()
	if err != nil {
		return err
	}

	log.Printf("Running schema migration on DB:%s", db.Migrator().CurrentDatabase())
	return adapter.Migrate(&models.Vulnerability{})
}
