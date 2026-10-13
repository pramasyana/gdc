// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config is the root application configuration.
type Config struct {
	App   App
	DB    DB
	Redis Redis
	JWT   JWT
}

// App holds application-level settings.
type App struct {
	Env  string
	Name string
	Port string
}

// IsProduction reports whether the application runs in production mode.
func (a App) IsProduction() bool { return a.Env == "production" }

// DB holds MySQL connection settings.
type DB struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DSN builds the MySQL data source name.
func (d DB) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

// Redis holds Redis connection settings.
type Redis struct {
	Addr     string
	Password string
	DB       int
}

// JWT holds token signing and lifetime settings.
type JWT struct {
	Secret     string
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// IdempotencyTTL returns the replay window for idempotent requests.
func (c Config) IdempotencyTTL() time.Duration {
	return viper.GetDuration("IDEMPOTENCY_TTL")
}

// Load reads configuration from an optional .env file and the environment.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()
	setDefaults()

	cfg := &Config{
		App: App{
			Env:  viper.GetString("APP_ENV"),
			Name: viper.GetString("APP_NAME"),
			Port: viper.GetString("APP_PORT"),
		},
		DB: DB{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetString("DB_PORT"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			Name:            viper.GetString("DB_NAME"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		Redis: Redis{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWT{
			Secret:     viper.GetString("JWT_SECRET"),
			Issuer:     viper.GetString("JWT_ISSUER"),
			AccessTTL:  viper.GetDuration("JWT_ACCESS_TTL"),
			RefreshTTL: viper.GetDuration("JWT_REFRESH_TTL"),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET must be set")
	}
	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_NAME", "gdc-task-api")
	viper.SetDefault("APP_PORT", "8080")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "3306")
	viper.SetDefault("DB_NAME", "gdc")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")

	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_ISSUER", "gdc-task-api")
	viper.SetDefault("JWT_ACCESS_TTL", "15m")
	viper.SetDefault("JWT_REFRESH_TTL", "168h")

	viper.SetDefault("IDEMPOTENCY_TTL", "24h")
}
