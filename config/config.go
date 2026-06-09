package config

import "os"

type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	JWTSecret  string
	Port       string
	LogGroup   string
	AWSRegion  string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "database-1.cluster-c0zimei407hh.us-east-1.rds.amazonaws.com"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "myapp_db"),
		DBUser:     getEnv("DB_USER", "myapp_user"),
		DBPassword: getEnv("DB_PASS", "nexthack-2026"),
		JWTSecret:  getEnv("JWT_SECRET", "secret"), // intentionally weak
		Port:       getEnv("PORT", "8080"),
		LogGroup:   getEnv("LOG_GROUP_NAME", "/dummy-be/app"),
		AWSRegion:  getEnv("AWS_REGION", "us-east-1"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
