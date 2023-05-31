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
	// SAMPLE MIGRATION
	// {
	// 	ID: "2023053000", // * Set to todays date, starting with 00 for first migration
	// 	Migrate: func(tx *gorm.DB) error {
	// 		// create a new ss_is_self_service bool column
	// 		if err := tx.Migrator().AddColumn(&Replication{}, "ss_is_self_service"); err != nil {
	// 			return err
	// 		}

	// 		// Update the values in the new column
	// 		if err := tx.Exec("UPDATE replications SET ss_is_self_service = is_self_service").Error; err != nil {
	// 			return err
	// 		}

	// 		// Remove the old column
	// 		if err := tx.Migrator().RenameColumn(&Replication{}, "is_self_service", "deprecated_is_self_service"); err != nil {
	// 			return err
	// 		}

	// 		return nil
	// 	},
	// 	Rollback: func(tx *gorm.DB) error {
	// 		// Rollback the migration by renaming the columns back to their original names
	// 		if err := tx.Migrator().RenameColumn(&Replication{}, "deprecated_is_self_service", "is_self_service"); err != nil {
	// 			return err
	// 		}

	// 		if err := tx.Migrator().RenameColumn(&Replication{}, "ss_is_self_service", "is_self_service"); err != nil {
	// 			return err
	// 		}

	// 		// Remove the new column
	// 		if err := tx.Migrator().DropColumn(&Replication{}, "ss_is_self_service"); err != nil {
	// 			return err
	// 		}

	// 		return nil
	// 	},
	// },
}
