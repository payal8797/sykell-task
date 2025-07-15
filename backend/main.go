package main

import (
	"github.com/gin-gonic/gin"
	"github.com/payal8797/sykell-task/backend/db"
)

func main() {
	db.InitDB()
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.Run(":8080") // Runs on http://localhost:8080
}
