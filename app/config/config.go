package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Config struct {
	DB   *sql.DB
	SMTP SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

var AppConfig *Config

func InitDB() {
	// Try remote database first, fallback to local
	var psqlInfo string

	// Check if LOCAL_DB environment variable is set
	if os.Getenv("LOCAL_DB") == "true" {
		psqlInfo = "host=localhost port=5432 user=postgres dbname=swadiq sslmode=disable"
		log.Println("Using local PostgreSQL database")
	} else {
		host := "129.80.85.203"
		port := 5432
		user := "imaad"
		password := "Ertdfgxc"
		dbname := "swadiq"

		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=60",
			host, port, user, password, dbname)
		log.Printf("Attempting to connect to remote database at %s:%d", host, port)
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to open database connection:", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Testing database connection...")
	if err = db.Ping(); err != nil {
		log.Printf("Database connection failed: %v", err)

		if os.Getenv("LOCAL_DB") != "true" {
			log.Println("\n=== DATABASE CONNECTION FAILED ===")
			log.Println("The remote database server is unreachable.")
			log.Println("\nTo use a local PostgreSQL database instead:")
			log.Println("1. Install PostgreSQL locally")
			log.Println("2. Create database: createdb swadiq")
			log.Println("3. Run schema: psql -d swadiq -f schema.sql")
			log.Println("4. Set environment variable: export LOCAL_DB=true")
			log.Println("5. Run the application again")
			log.Println("\nOr check the remote database connection:")
			log.Println("- Verify IP address: 129.80.199.242")
			log.Println("- Check network connectivity")
			log.Println("- Ensure database 'swadiq' exists")
			log.Println("- Verify user 'imaad' has access")
		}

		log.Fatal("Cannot establish database connection")
	}

	AppConfig = &Config{
		DB: db,
		SMTP: SMTPConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			Username: "swadiqjuniorschools@gmail.com",
			Password: "varn boqq brqq ftjv",
			From:     "swadiqjuniorschools@gmail.com",
		},
	}
	log.Println("Database connected successfully")
	log.Println("Email configuration initialized")
}

func GetDB() *sql.DB {
	return AppConfig.DB
}
