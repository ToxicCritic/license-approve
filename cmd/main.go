package main

import (
	"LicenseApp/pkg/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()
	db.Migrate()
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Server is running!"})
	})

	// Проверка лицензии
	r.GET("/license/check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "License is valid",
		})
	})

	// Создание лицензии
	r.POST("/license/create", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"status":  "success",
			"message": "License created",
		})
	})

	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
