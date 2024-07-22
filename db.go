package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB = nil

func dbConn() *gorm.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	if db == nil {
		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
	}
	return db
}

type UserType string

const (
	UserTypeApplicant UserType = "applicant"
	UserTypeAdmin     UserType = "admin"
)

type User struct {
	gorm.Model
	Name            string   `gorm:"size:255;not null"`
	Email           string   `gorm:"size:255;not null;unique"`
	Address         string   `gorm:"size:255"`
	UserType        UserType `gorm:"size:20;not null"`
	PasswordHash    string   `gorm:"size:255;not null"`
	ProfileHeadline string   `gorm:"size:255"`
	Profile         Profile  `gorm:"foreignKey:UserID"`
}

type Profile struct {
	gorm.Model
	UserID            uint   `gorm:"not null"`
	ResumeFileAddress string `gorm:"size:255"`
	Skills            string `gorm:"type:text"`
	Education         string `gorm:"type:text"`
	Experience        string `gorm:"type:text"`
	Name              string `gorm:"size:255"`
	Email             string `gorm:"size:255"`
	Phone             string `gorm:"size:20"`
}

type Job struct {
	gorm.Model
	Title             string    `gorm:"size:255;not null"`
	Description       string    `gorm:"type:text"`
	PostedOn          time.Time `gorm:"not null"`
	TotalApplications int       `gorm:"default:0"`
	CompanyName       string    `gorm:"size:255;not null"`
	PostedByID        uint      `gorm:"not null"`
	PostedBy          User      `gorm:"foreignKey:PostedByID"`
	Applicants        []User    `gorm:"many2many:job_applications;"`
}
