package core

import (
	"time"

	"gorm.io/gorm"
)

func Watch(db *gorm.DB) {
	go watch(db)
}

func watch(db *gorm.DB) {
	for {

		time.Sleep(10 * time.Second)
	}
}
