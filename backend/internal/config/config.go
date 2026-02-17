package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Port                  string
	DBPath                string
	MigrationDir          string
	JWTSecret             string
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	PasswordResetTokenTTL time.Duration
	AppEnv                string
}

func Load() Config {
	return Config{
		Port:                  getEnv("APP_PORT", "8080"),
		DBPath:                getEnv("DB_PATH", defaultDBPath()),
		MigrationDir:          getEnv("MIGRATIONS_DIR", defaultMigrationDir()),
		JWTSecret:             getEnv("JWT_SECRET", "dev-secret-change-me"),
		AccessTokenTTL:        time.Duration(getEnvAsInt("ACCESS_TOKEN_MIN", 15)) * time.Minute,
		RefreshTokenTTL:       time.Duration(getEnvAsInt("REFRESH_TOKEN_HOURS", 24*7)) * time.Hour,
		PasswordResetTokenTTL: time.Duration(getEnvAsInt("PASSWORD_RESET_MIN", 30)) * time.Minute,
		AppEnv:                getEnv("APP_ENV", "dev"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func defaultMigrationDir() string {
	if dirExists("./migrations") {
		return "./migrations"
	}
	return "./backend/migrations"
}

func defaultDBPath() string {
	if dirExists("./migrations") {
		return "./data/app.db"
	}
	return "./backend/data/app.db"
}

func dirExists(path string) bool {
	info, err := os.Stat(filepath.Clean(path))
	if err != nil {
		return false
	}
	return info.IsDir()
}
