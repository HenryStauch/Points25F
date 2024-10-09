package src

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	db, err := gorm.Open(sqlite.Open("points24F.db"), &gorm.Config{})
	if err != nil {
		panic("ERROR: Could not open database file")
	}

	pledge_err := db.AutoMigrate(&Pledge{})
	brother_err := db.AutoMigrate(&Brother{})
	points_err := db.AutoMigrate(&Point{})
	if pledge_err != nil || brother_err != nil || points_err != nil {
		panic("ERROR: Could not migrate pledge data")
	}
	DB = db

	var dev Brother = Brother{
		Name:      "Rishav",
		BrotherId: "55200205",
		IsAdmin:   true,
		IsTimeout: false,
	}
	_ = DB.FirstOrCreate(&dev)
}
