package main

import (
	"fmt"
	"net/http"
	"rishavc/points/src"

	"github.com/gin-gonic/gin"
	// "github.com/joho/godotenv"
)

func main() {
	fmt.Println("Beep beep starting points bot!")

	// err := godotenv.Load()
	// if err != nil {
	// 	panic("ERROR: Could not load .env file")
	// }

	src.ConnectDatabase()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("", src.ReceiveChat)
	r.Run()
}
