package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
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

func uploadResume(c *gin.Context) {
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

	log.Println(user.UserType)
	if user.UserType != "applicant" {
		c.JSON(403, gin.H{
			"error": "Only applicants can access this endpoint",
		})
		return
	}

	fileHeader, err := c.FormFile("resume")
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Resume file is required",
		})
		return
	}

	resumeAPIKey := os.Getenv("RESUME_API_KEY")
	if resumeAPIKey == "" {
		c.JSON(500, gin.H{
			"error": "Can not process resume",
		})
		log.Println("RESUME_API_KEY is not set")
		return
	}

	if fileHeader.Header.Get("Content-Type") != "application/pdf" && fileHeader.Header.Get("Content-Type") != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		c.JSON(400, gin.H{
			"error": "Resume file should be a pdf or docx",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error reading file",
		})
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error reading file",
		})
		return
	}

	apiUrl := "https://api.apilayer.com/resume_parser/upload"
	apikey := os.Getenv("RESUME_API_KEY")

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(fileBytes))
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error uploading resume",
		})
		return
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("apikey", apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error uploading resume",
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error reading response",
		})
		return
	}

	var result interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error parsing response",
		})
		return
	}

	skills, _ := json.Marshal(result.(map[string]interface{})["skills"])
	education, _ := json.Marshal(result.(map[string]interface{})["education"])
	experience, _ := json.Marshal(result.(map[string]interface{})["experience"])
	name := user.Name
	email := user.Email
	phone := result.(map[string]interface{})["phone"]


	db.Model(&user).Association("Profile").Append(&Profile{
		UserID: user.ID,
		ResumeFileAddress: "",
		Skills: string(skills),
		Education: string(education),
		Experience: string(experience),
		Name: name,
		Email: email,
		Phone: phone.(string),
	})

	c.JSON(200, gin.H{
		"message": "Resume uploaded successfully",
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

func getJob(c *gin.Context) {
	jobId := c.Param("job_id")

	db := dbConn()

	var job Job
	tx := db.Model(&Job{}).Preload("Applicants").Preload("PostedBy").Preload("Applicants.Profile").First(&job, jobId)
	if tx.Error != nil {
		c.JSON(404, gin.H{
			"error": "Job not found",
		})
		return
	}
	

	c.JSON(200, job)
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

	tx := db.Where("user_type = ?", "applicant").Preload("Profile").Find(&applicants)
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

	tx := db.Where("id = ? AND user_type = ?", id, "applicant").Preload("Profile").First(&applicant)
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

func applyJob(c *gin.Context) {
	jobId := c.Query("job_id")
	if jobId == "" {
		c.JSON(400, gin.H{
			"error": "job_id is a required parameter",
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

	applicant, err := getUserFromToken(tokenString)
	if err != nil {
		c.JSON(401, gin.H{
			"error": "Invalid token",
		})
		return
	}

	if applicant.UserType != "applicant" {
		c.JSON(403, gin.H{
			"error": "Only applicants can apply for jobs",
		})
		return
	}

	db := dbConn()

	var job Job
	tx := db.Where("id = ?", jobId).First(&job)
	if tx.Error != nil {
		c.JSON(404, gin.H{
			"error": "Job not found",
		})
		return
	}
	
	tx = db.Begin()
	
	err = db.Model(&job).Association("Applicants").Append(&applicant)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error": "Error applying for job",
		})
		return
	}
	
	up := db.Model(&Job{}).Where("id = ?", job.ID).Update("total_applications", job.TotalApplications+1)
	if up.Error != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error": "Error applying for job",
		})
		return
	}

	tx.Commit()

	c.JSON(200, gin.H{
		"message": "Applied for job successfully",
	})
}