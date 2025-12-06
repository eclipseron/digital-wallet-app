package conf

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupDB() *gorm.DB {
	godotenv.Load()
	opt := gorm.Config{
		DefaultTransactionTimeout: 20 * time.Second,
		DefaultContextTimeout:     60 * time.Second,
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &opt)
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	if err := db.Exec("SELECT 1").Error; err != nil {
		log.Fatal("failed to connect database:", err)
	}
	log.Println("DB connection initiated")

	return db
}
