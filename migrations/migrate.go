package migrations

import (
	"log"

	"github.com/eclipseron/digital-wallet-app/models"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Fatal("failed to enable uuid extension: ", err)
	}

	err := db.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Transactions{},
	)
	if err != nil {
		log.Fatal("migration failed:", err)
	}
	log.Println("migration success")
}
