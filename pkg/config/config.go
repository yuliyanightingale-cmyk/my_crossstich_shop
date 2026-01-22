package config

import (
	"my_crossstich_shop/pkg/models"
	"os"

	"github.com/joho/godotenv"
)

func New() (*models.Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	cfg := &models.Config{
		DbHost:     os.Getenv("DB_HOST"),
		DbPort:     os.Getenv("DB_PORT"),
		DbUser:     os.Getenv("DB_USER"),
		DbPassword: os.Getenv("DB_PASSWORD"),
		DbName:     os.Getenv("DB_NANE"),
		DbSslmode:  os.Getenv("DB_SSLMODE"),
	}

	return cfg, nil
}
