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

type User struct {
	Id         string `json:"id"`
	UserId     string `json:"user_id"`
	Nickname   string `json:"nickname"`
	Name       string `json:"name"`
	ImageUrl   string `json:"image_url"`
	Muted      bool   `json:"muted"`
	Autokicked bool   `json:"autokicked"`
	Roles      []any  `json:"roles"`
}

// {"user_id":"55200205","nickname":"Rishav Chakravarty","image_url":"https://i.groupme.com/979x979.jpeg.ad5c263dd9b648fc8cec4ad5f4a1a612","id":"1019915005","muted":false,"autokicked":false,"roles":["admin"],"name":"Rishav Chakravarty"}
