package db

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func BaselineSchema(tx *gorm.DB) error {
	log.Debugf("first run: initializing database schema")
	err := tx.AutoMigrate(&Provider{}, &Dataset{}, &Content{}, &Wallet{}, &ProviderAllowedDatasets{}, &WalletDatasets{}, &Replication{})

	if err != nil {
		log.Fatalf("error initializing database: %s", err)
	}
	return nil
}

var Migrations []*gormigrate.Migration = []*gormigrate.Migration{
	// Move isSelfService column inside SelfService struct
	{
		ID: "20230530000",
		Migrate: func(tx *gorm.DB) error {
			// ! Column has already been created as it's included in the baseline

			// Update the values in the new column
			if err := tx.Exec("UPDATE replications SET ss_is_self_service = is_self_service").Error; err != nil {
				return err
			}

			// Remove the old column
			if err := tx.Migrator().RenameColumn(&Replication{}, "is_self_service", "deprecated_is_self_service"); err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Rollback the migration by renaming the columns back to their original names
			if err := tx.Migrator().RenameColumn(&Replication{}, "deprecated_is_self_service", "is_self_service"); err != nil {
				return err
			}

			if err := tx.Migrator().RenameColumn(&Replication{}, "ss_is_self_service", "is_self_service"); err != nil {
				return err
			}

			// Remove the new column
			if err := tx.Migrator().DropColumn(&Replication{}, "ss_is_self_service"); err != nil {
				return err
			}

			return nil
		},
	},
}
