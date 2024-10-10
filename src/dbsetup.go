package src

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE"),
		os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("ERROR: Database connection refused")
	} else {
		fmt.Println("Database connection successful")
	}

	pledge_err := db.AutoMigrate(&Pledge{})
	brother_err := db.AutoMigrate(&Brother{})
	points_err := db.AutoMigrate(&Point{})
	if pledge_err != nil || brother_err != nil || points_err != nil {
		panic("ERROR: Could not migrate database data")
	}
	DB = db

	// Add me as a permanent admin
	var dev Brother = Brother{
		Name:      "Rishav",
		BrotherId: "55200205",
		IsAdmin:   true,
		IsTimeout: false,
	}
	_ = DB.FirstOrCreate(&dev)
}
