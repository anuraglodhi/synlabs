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
	db.AutoMigrate(&User{}, &Profile{}, &Job{})

	r.POST("/signup", signup)
	r.POST("/login", login)
	r.POST("/uploadResume", uploadResume)
	r.POST("/admin/job", createJob)
	r.GET("/admin/job/:job_id", getJob)
	r.GET("/admin/applicants", getApplicants)
	r.GET("/admin/applicant/:applicant_id", getApplicant)
	r.GET("/jobs", getJobs)
	r.GET("/jobs/apply", applyJob)

	r.Run()
}
