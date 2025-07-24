package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var once sync.Once

// LoadEnv loads environment variables from .env file
// This is safe to call multiple times
func LoadEnv() {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: .env file not found, using system environment variables")
		}
	})
}

// Helper function to get environment variable with a default value
func GetEnv(key, defaultValue string) string {
	LoadEnv() // Ensure .env is loaded
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvOrPanic(key string) string {
	LoadEnv() // Ensure .env is loaded
	value := os.Getenv(key)
	if value == "" {
		panic("Environment variable not set: " + key)
	}
	return value
}

// GetEnvBool gets environment variable as boolean with default
func GetEnvBool(key string, defaultValue bool) bool {
	LoadEnv()
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// IsDebugEnabled checks if debug logging is enabled
func IsDebugEnabled() bool {
	return GetEnvBool("DEBUG", false)
}

// DebugLog prints debug message only if DEBUG=true
func DebugLog(format string, args ...interface{}) {
	if IsDebugEnabled() {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// DebugLogf is an alias for DebugLog (formatted)
func DebugLogf(format string, args ...interface{}) {
	DebugLog(format, args...)
}
