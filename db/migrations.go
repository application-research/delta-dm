package db

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
)

var Migrations []*gormigrate.Migration = []*gormigrate.Migration{
	// Move isSelfService column inside SelfService struct
	// {
	// 	ID: "20230530000",
	// 	Migrate: func(tx *gorm.DB) error {
	// 		tx.Migrator().RenameColumn(&SelfService{}, "is_self_service", "is_self_service_old")
	// 		// return tx.AutoMigrate(&Dataset{})
	// 	},
	// 	Rollback: func(tx *gorm.DB) error {
	// 		return tx.Migrator().DropTable("datasets")
	// 	},
	// },
}
