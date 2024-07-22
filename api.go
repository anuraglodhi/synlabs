package main

import (
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