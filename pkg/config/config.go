package config

import (
	"log"
	"os"
)

type Config struct {
	Env               string
	ListenAddr        string
	ControlListenAddr string
	LogLevel          string
	DBDSN             string
}

func Load() Config {
	cfg := Config{
		Env:               getEnv("APP_ENV", "local"),
		ListenAddr:        getEnv("GATEWAY_LISTEN_ADDR", ":8080"),
		ControlListenAddr: getEnv("CONTROL_LISTEN_ADDR", ":8081"),
		LogLevel:          getEnv("LOG_LEVEL", "debug"),
		DBDSN:             getEnv("DB_DSN", "file:nexus.db?_foreign_keys=on"),
	}

	log.Printf("config loaded: %+v\n", cfg)

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
