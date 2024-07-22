package main

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

func createToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":  user.ID,
			"exp": time.Now().Add(time.Hour * 72).Unix(),
		})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func getUserFromToken(tokenString string) (User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return User{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return User{}, err
	}

	var user User
	db := dbConn()
	tx := db.First(&user, int(claims["id"].(float64)))
	if tx.Error != nil {
		return User{}, tx.Error
	}

	return user, nil
}
