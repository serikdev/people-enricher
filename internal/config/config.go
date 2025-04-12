package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerCfg
	DBConfig DBCfg
	Logger   LoggerCfg
}

type ServerCfg struct {
	Port      string
	Host      string
	ReadTime  time.Duration
	WriteTime time.Duration
}

type DBCfg struct {
	Port           string
	Host           string
	User           string
	Password       string
	DBName         string
	SSLMode        string
	MaxConnection  int
	IdleConnection int
	ConnectLife    time.Duration
}

type LoggerCfg struct {
	Level string
}

func LoadCfg(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("failed load env file: %w", err)
	}
	readTimeout, err := time.ParseDuration(getEnv("SERVER_READ_TIME", "60s"))
	if err != nil {
		return nil, fmt.Errorf("parsing read timeout: %w", err)
	}
	writeTimeout, err := time.ParseDuration(getEnv("SERVER_WRITE_TIME", "60s"))
	if err != nil {
		return nil, fmt.Errorf("parsing write timeout: %w", err)
	}
	maxConnection, err := strconv.Atoi(getEnv("DB_MAX_CONNECTIONS", "10"))
	if err != nil {
		return nil, fmt.Errorf("parsing max connection: %w", err)
	}
	maxIdleConnection, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNECTIONS", "5"))
	if err != nil {
		return nil, fmt.Errorf("parsing max idle connction: %w", err)
	}
	connLifetime, err := time.ParseDuration(getEnv("DB_MAX_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("parsing connection lifetime: %w", err)
	}

	return &Config{
		Server: ServerCfg{
			Port:      getEnv("SERVER_PORT", "8080"),
			Host:      getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTime:  readTimeout,
			WriteTime: writeTimeout,
		},
		DBConfig: DBCfg{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnv("DB_PORT", "5432"),
			User:           getEnv("DB_USER", "postgres"),
			Password:       getEnv("DB_PASSWORD", ""),
			DBName:         getEnv("DB_NAME", "peopleEnricher"),
			SSLMode:        getEnv("DB_SSL_MODE", "disable"),
			MaxConnection:  maxConnection,
			IdleConnection: maxIdleConnection,
			ConnectLife:    connLifetime,
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
