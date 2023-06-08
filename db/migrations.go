package db

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// If this runs, it means the database is empty. No migrations will be applied on top of it, as this sets up the database from scratch so it starts out "up to date"
func BaselineSchema(tx *gorm.DB) error {
	log.Debugf("first run: initializing database schema")
	err := tx.AutoMigrate(&Provider{}, &Dataset{}, &Content{}, &Wallet{}, &ReplicationProfile{}, &WalletDatasets{}, &Replication{})

	if err != nil {
		log.Fatalf("error initializing database: %s", err)
	}
	return nil
}

var Migrations []*gormigrate.Migration = []*gormigrate.Migration{
	{
		ID: "2023060800", // Set to todays date, starting with 00 for first migration
		Migrate: func(tx *gorm.DB) error {
			return tx.Migrator().AddColumn(&Content{}, "ContentLocation")
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropColumn(&Content{}, "ContentLocation")
		},
	},
}
