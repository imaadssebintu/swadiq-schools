package main

import (
	"fmt"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
)

func main() {
	// Initialize database connection
	config.InitDB()
	db := config.GetDB()
	if db == nil {
		fmt.Println("Failed to connect to database")
		return
	}

	// Create user
	user := &models.User{
		FirstName: "Imaad",
		LastName:  "Dean",
		Email:     "imaad.dean@gmail.com",
		Password:  "Ertdfgx@0",
	}

	err := database.CreateTeacher(db, user)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}

	fmt.Printf("User created successfully: %s %s (%s)\n", user.FirstName, user.LastName, user.Email)
}
