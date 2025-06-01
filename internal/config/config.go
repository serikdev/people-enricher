package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConfig    DBCfg
	Logger      LoggerCfg
	ExternalAPI ExternalAPIConfig
}

type DBCfg struct {
	Port     string
	Host     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}
type ExternalAPIConfig struct {
	AgifyURL       string
	GenderizeURL   string
	NationalizeURL string
}
type LoggerCfg struct {
	Level string
}

func LoadCfg(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("failed load env file: %w", err)
	}

	return &Config{
		DBConfig: DBCfg{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "peopleEnricher"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		ExternalAPI: ExternalAPIConfig{
			AgifyURL:       os.Getenv("AGIFY_API_URL"),
			GenderizeURL:   os.Getenv("GENDERIZE_API_URL"),
			NationalizeURL: os.Getenv("NATIONALIZE_API_URL"),
		},
		Logger: LoggerCfg{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
