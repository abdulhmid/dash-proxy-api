package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Admin    AdminConfig    `mapstructure:"admin"`
}

type ServerConfig struct {
	Port  int    `mapstructure:"port"`
	Proxy string `mapstructure:"proxy"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Mongo    MongoConfig    `mapstructure:"mongo"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type MongoConfig struct {
	URI    string `mapstructure:"uri"`
	DBName string `mapstructure:"dbname"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpiryHour int    `mapstructure:"expiry_hour"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return strings.ToLower(v) == "true" || v == "1"
	}
	return fallback
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("warn: config file not found, using env vars: %v", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:  8080,
			Proxy: env("SERVER_PROXY", ""),
		},
		Database: DatabaseConfig{
			Postgres: PostgresConfig{
				Host:     env("DATABASE_POSTGRES_HOST", "localhost-noset"),
				Port:     env("DATABASE_POSTGRES_PORT", "5432"),
				User:     env("DATABASE_POSTGRES_USER", "postgres"),
				Password: env("DATABASE_POSTGRES_PASSWORD", "postgres"),
				DBName:   env("DATABASE_POSTGRES_DBNAME", "api_source_proxy"),
				SSLMode:  env("DATABASE_POSTGRES_SSLMODE", "disable"),
			},
			Mongo: MongoConfig{
				URI:    env("DATABASE_MONGO_URI", "mongodb://localhost:27017"),
				DBName: env("DATABASE_MONGO_DBNAME", "api_source_proxy"),
			},
		},
		JWT: JWTConfig{
			Secret:     env("JWT_SECRET", "change-me"),
			ExpiryHour: envInt("JWT_EXPIRY_HOUR", 24),
		},
		Admin: AdminConfig{
			Username: env("ADMIN_USERNAME", "admin"),
			Password: env("ADMIN_PASSWORD", "admin123"),
			Email:    env("ADMIN_EMAIL", "admin@example.com"),
		},
	}

	if err := viper.Unmarshal(cfg); err != nil {
		log.Printf("warn: failed to parse config file, using defaults: %v", err)
	}

	return cfg
}
