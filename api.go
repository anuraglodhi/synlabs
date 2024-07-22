package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func signup(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	userType := c.PostForm("userType")
	profileHeadline := c.PostForm("profileHeadline")
	address := c.PostForm("address")

	if name == "" || email == "" || password == "" || userType == "" {
		c.JSON(400, gin.H{
			"error": "name, email, password, and userType are required fields",
		})
		return
	}

	if userType != "applicant" && userType != "admin" {
		c.JSON(400, gin.H{
			"error": "userType should be either applicant or admin",
		})
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error hashing password",
		})
		return
	}

	db := dbConn()

	tx := db.Where("email = ?", email).First(&User{})
	if tx.RowsAffected > 0 {
		c.JSON(400, gin.H{
			"error": "User with that email already exists",
		})
		return
	}

	tx = db.Create(&User{
		Name:            name,
		Email:           email,
		Address:         address,
		UserType:        UserType(userType),
		PasswordHash:    hashedPassword,
		ProfileHeadline: profileHeadline,
	})

	if tx.Error != nil {
		c.JSON(500, gin.H{
			"error": "Error creating user",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "User created successfully",
	})
}

func login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.JSON(400, gin.H{
			"error": "email and password are required fields",
		})
		return
	}

	db := dbConn()

	var user User
	tx := db.Where("email = ?", email).First(&user)

	if tx.Error != nil {
		c.JSON(404, gin.H{
			"error": "User not found",
		})
		return
	}

	if !checkPasswordHash(password, user.PasswordHash) {
		c.JSON(401, gin.H{
			"error": "Invalid password",
		})
		return
	}

	token, err := createToken(user)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error creating token",
		})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}

func createJob(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	companyName := c.PostForm("companyName")

	if title == "" || description == "" || companyName == "" {
		c.JSON(400, gin.H{
			"error": "title, description, and companyName are required fields",
		})
		return
	}

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(401, gin.H{
			"error": "Authorization header is required",
		})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	poster, err := getUserFromToken(tokenString)
	if err != nil {
		c.JSON(401, gin.H{
			"error": "Invalid token",
		})
		return
	}

	if poster.UserType != "admin" {
		c.JSON(403, gin.H{
			"error": "Only admins can access this endpoint",
		})
		return
	}

	tx := db.Create(&Job{
		Title:             title,
		Description:       description,
		CompanyName:       companyName,
		PostedByID:        poster.ID,
		PostedOn:          time.Now(),
		TotalApplications: 0,
	})
	if tx.Error != nil {
		c.JSON(500, gin.H{
			"error": "Error creating job",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Job created successfully",
	})
}

func getApplicants(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(401, gin.H{
			"error": "Authorization header is required",
		})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := getUserFromToken(tokenString)
	if err != nil {
		c.JSON(401, gin.H{
			"error": "Invalid token",
		})
		return
	}

	if user.UserType != "admin" {
		c.JSON(403, gin.H{
			"error": "Only admins can access this endpoint",
		})
		return
	}

	db := dbConn()

	var applicants []User

	tx := db.Where("user_type = ?", "applicant").Find(&applicants)
	if tx.Error != nil {
		c.JSON(500, gin.H{
			"error": "Error fetching applicants",
		})
		return
	}

	c.JSON(200, gin.H{
		"applicants": applicants,
	})
}

func getApplicant(c *gin.Context) {
	id := c.Param("applicant_id")
	if id == "" {
		c.JSON(400, gin.H{
			"error": "applicant_id is a required parameter",
		})
	}

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(401, gin.H{
			"error": "Authorization header is required",
		})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := getUserFromToken(tokenString)
	if err != nil {
		c.JSON(401, gin.H{
			"error": "Invalid token",
		})
		return
	}

	if user.UserType != "admin" {
		c.JSON(403, gin.H{
			"error": "Only admins can access this endpoint",
		})
		return
	}

	db := dbConn()

	var applicant User

	tx := db.Where("id = ? AND user_type = ?", id, "applicant").First(&applicant)
	if tx.Error != nil {
		c.JSON(404, gin.H{
			"error": "Applicant not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"applicant": applicant,
	})
}

func getJobs(c *gin.Context) {
	db := dbConn()

	var jobs []Job
	
	// Using new structs to avoid sending sensitive information

	type SafeUser struct {
		ID    uint
		Email string
		Name  string
	}

	type JobWithoutApplicants struct {
		ID                uint
		Title             string
		Description       string
		CompanyName       string
		PostedByID        uint
		PostedOn          time.Time
		TotalApplications int
		PostedBy          SafeUser
	}

	tx := db.Model(&Job{}).Preload("PostedBy").Find(&jobs)
	if tx.Error != nil {
		c.JSON(500, gin.H{
			"error": "Error fetching jobs",
		})
		return
	}

	var jobsWithoutApplicants []JobWithoutApplicants
	for _, job := range jobs {
		jobsWithoutApplicants = append(jobsWithoutApplicants, JobWithoutApplicants{
			ID:                job.ID,
			Title:             job.Title,
			Description:       job.Description,
			CompanyName:       job.CompanyName,
			PostedByID:        job.PostedByID,
			PostedOn:          job.PostedOn,
			TotalApplications: job.TotalApplications,
			PostedBy: SafeUser{
				ID:    job.PostedBy.ID,
				Email: job.PostedBy.Email,
				Name:  job.PostedBy.Name,
			},
		})
	}

	c.JSON(200, gin.H{
		"jobs": jobsWithoutApplicants,
	})
}
