package src

import "gorm.io/gorm"

type Pledge struct {
	gorm.Model
	Name     string
	PledgeId string
	Points   int
}

type Brother struct {
	gorm.Model
	Name      string
	BrotherId string // Assigned by GroupMe UserId
	IsAdmin   bool
	IsTimeout bool // in case someone is being a real tool
}

type Point struct {
	gorm.Model
	PointsGiven int
	PledgeId    uint
}

// User model only used for adding users to the DB at term start
type User struct {
	UserId   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
}
