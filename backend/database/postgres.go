package postgres

import (
	"backend/database/models"
	"backend/pkg/utils"
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
having env variables as global variables might not be the best practice
had to make a simpleton for loading .env file each call
*/
var db_name = utils.GetEnvOrPanic("DB_NAME")

// var db_user = utils.GetEnvOrPanic("DB_USER")
// var db_user_password = utils.GetEnvOrPanic("DB_USER_PASSWORD")
var db_postgres_password = utils.GetEnv("DB_POSTGRES_PASSWORD", "")
var db_host = utils.GetEnvOrPanic("DB_HOST")
var db_port = utils.GetEnvOrPanic("DB_PORT")

var db *gorm.DB

type Queries struct {
	db *gorm.DB
}

func createDatabaseIfNotExists() error {
	// Connect to default postgres database to create our app database
	postgresURL := fmt.Sprintf("host=%s user=postgres password=%s dbname=postgres port=%s sslmode=disable",
		utils.GetEnvOrPanic("DB_HOST"),
		utils.GetEnvOrPanic("DB_POSTGRES_PASSWORD"),
		utils.GetEnvOrPanic("DB_PORT"))

	db, err := gorm.Open(postgres.Open(postgresURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}

	dbName := utils.GetEnvOrPanic("DB_NAME")

	// Check if database exists first
	var exists bool
	checkQuery := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s')", dbName)
	result := db.Raw(checkQuery).Scan(&exists)
	if result.Error != nil {
		return fmt.Errorf("failed to check if database exists: %w", result.Error)
	}

	if exists {
		log.Printf("Database %s already exists\n", dbName)
		return nil
	}

	// Create database if it doesn't exist
	createResult := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if createResult.Error != nil {
		// Check if error is "permission denied" - warn but continue
		if strings.Contains(createResult.Error.Error(), "permission denied") {
			fmt.Printf("Warning: Cannot create database %s (permission denied). Please create it manually:\n", dbName)
			fmt.Printf("  psql -U postgres -c \"CREATE DATABASE %s OWNER %s;\"\n", dbName, utils.GetEnvOrPanic("DB_USER"))
			return nil // Don't fail, assume database exists
		}
		return fmt.Errorf("failed to create database %s: %w", dbName, createResult.Error)
	}

	fmt.Printf("Database %s created successfully\n", dbName)
	return nil
}

func InitDb() error {
	if err := createDatabaseIfNotExists(); err != nil {
		return err
	}

	dsn := fmt.Sprintf("host=%s user=postgres password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Tallinn",
		db_host, db_postgres_password, db_name, db_port)

	d, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return err
	}
	db = d

	err = db.AutoMigrate(&models.Sensor{}, &models.User{}, &models.SensorReading{})
	if err != nil {
		fmt.Printf("Failed to auto migrate models: %v\n", err)
		return err
	}

	fmt.Println("Connected to PostgreSQL")
	return nil
}

func New() *Queries {
	return &Queries{db: db}
}

func DB() *gorm.DB {
	return db
}
