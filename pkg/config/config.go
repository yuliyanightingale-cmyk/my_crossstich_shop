package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassword string
	DbName     string
	DbSslmode  string
}

func New() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DbHost:     os.Getenv("DB_HOST"),
		DbPort:     os.Getenv("DB_PORT"),
		DbUser:     os.Getenv("DB_USER"),
		DbPassword: os.Getenv("DB_PASSWORD"),
		DbName:     os.Getenv("DB_NANE"),
		DbSslmode:  os.Getenv("DB_SSLMODE"),
	}

	return cfg, nil
}
