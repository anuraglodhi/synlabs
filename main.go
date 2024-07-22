package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	db := dbConn()
	db.AutoMigrate(&User{}, &Profile{}, &Job{}, &JobApplication{})

	r.POST("/signup", signup)
	r.POST("/login", login)
	r.GET("/admin/applicants", getApplicants)
	r.GET("/admin/applicant/:applicant_id", getApplicant)

	r.Run()
}
