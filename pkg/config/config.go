package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppName string
	AppPort string
	AppEnv  string

	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret  string
	JWTIssuer  string
	JWTJWKSURL string

	APIKey string

	CORSAllowedOrigins string
	RateLimitEnabled   bool

	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string
	AppURL   string
}

func Load() (*Config, error) {
	redisDB, _ := strconv.Atoi(env("REDIS_DB", "0"))
	rateLimitEnabled, _ := strconv.ParseBool(env("RATE_LIMIT_ENABLED", "true"))

	cfg := &Config{
		AppName: env("APP_NAME", "mirandaclin"),
		AppPort: env("APP_PORT", "8080"),
		AppEnv:  env("APP_ENV", "development"),

		DBHost:    env("DB_HOST", "localhost"),
		DBPort:    env("DB_PORT", "5432"),
		DBUser:    env("DB_USER", "postgres"),
		DBPass:    env("DB_PASS", ""),
		DBName:    env("DB_NAME", "mirandaclin"),
		DBSSLMode: env("DB_SSLMODE", "disable"),

		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		JWTSecret:  env("JWT_SECRET", ""),
		JWTIssuer:  env("JWT_ISSUER", ""),
		JWTJWKSURL: env("JWT_JWKS_URL", ""),

		APIKey: env("API_KEY", ""),

		CORSAllowedOrigins: env("CORS_ALLOWED_ORIGINS", ""),
		RateLimitEnabled:   rateLimitEnabled,

		SMTPHost: env("SMTP_HOST", ""),
		SMTPPort: env("SMTP_PORT", "587"),
		SMTPUser: env("SMTP_USER", ""),
		SMTPPass: env("SMTP_PASS", ""),
		SMTPFrom: env("SMTP_FROM", ""),
		AppURL:   env("APP_URL", "http://localhost:3000"),
	}

	if cfg.AppEnv == "production" && cfg.DBSSLMode == "disable" {
		return nil, fmt.Errorf("DB_SSLMODE=disable é proibido em production")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET é obrigatório")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API_KEY é obrigatório")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=America/Sao_Paulo",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName, c.DBSSLMode,
	)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
