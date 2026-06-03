package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	DatabaseURL  string
	JWTSecret    []byte
	JWTTokenTTL  time.Duration
	TemplatesDir string
	StaticDir    string
	Addr         string
	CookieTTL    time.Duration
}

func Load() *Config {
	return &Config{
		DatabaseURL:  mustGetEnv("DATABASE_URL"),
		JWTSecret:    []byte(mustGetEnv("JWT_SECRET")),
		JWTTokenTTL:  24 * time.Hour,
		TemplatesDir: getEnv("TEMPLATES_DIR", "web/templates"),
		StaticDir:    getEnv("STATIC_DIR", "web/static"),
		Addr:         getEnv("ADDR", "localhost:8080"),
		CookieTTL:    time.Hour,
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
