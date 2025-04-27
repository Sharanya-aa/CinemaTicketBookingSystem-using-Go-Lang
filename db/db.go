package db

import (
	"fmt"
	"log"
	"os"
	"sync"

	"cinemabooking/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db   *gorm.DB
	once sync.Once
)

func InitDB() *gorm.DB {
	once.Do(func() {
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbHost := os.Getenv("DB_HOST")
		dbName := os.Getenv("DB_NAME")

		if dbUser == "" {
			dbUser = "root"
		}
		if dbPass == "" {
			dbPass = "@Kg200347"
		}
		if dbHost == "" {
			dbHost = "localhost:3306"
		}
		if dbName == "" {
			dbName = "cinema_booking"
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPass, dbHost, dbName)

		var err error
		db, err = gorm.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Auto migrate the schema
		db.AutoMigrate(&models.Movie{}, &models.Show{}, &models.Seat{}, &models.Booking{}, &models.User{})

		// Enable logging
		db.LogMode(true)
	})

	return db
}

func GetDB() *gorm.DB {
	if db == nil {
		return InitDB()
	}
	return db
}
