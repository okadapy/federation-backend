package main

import (
	"fmt"
	"os"
)

type DBConfig struct {
	Driver string
	Host   string
	Port   string
	User   string
	Pass   string
	Name   string
}

type ServerConfig struct {
	Port string
	Host string
}

type AppConfig struct {
	FileStoragePath string
}

type Config struct {
	DB     DBConfig
	Server ServerConfig
	App    AppConfig
}

func NewConfig() *Config {
	return &Config{
		DB: DBConfig{
			Driver: getEnv("DB_DRIVER", "postgres"),
			Host:   getEnv("DB_HOST", "db"),
			Port:   getEnv("DB_PORT", "5432"),
			User:   getEnv("DB_USER", "postgres"),
			Pass:   getEnv("DB_PASS", "postgres"),
			Name:   getEnv("DB_NAME", "postgres"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		App: AppConfig{
			FileStoragePath: getEnv("APP_FILE_STORAGE_PATH", "./files"),
		},
	}
}

func (db *DBConfig) GetDBUri() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		db.Driver,
		db.User,
		db.Pass,
		db.Host,
		db.Port,
		db.Name,
	)
}

func (server *ServerConfig) GetHostURL() string {
	return fmt.Sprintf("%s:%s", server.Host, server.Port)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
