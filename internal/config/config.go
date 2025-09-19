package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Database   DatabaseConfig   `mapstructure:"database"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	Redis      RedisConfig      `mapstructure:"redis"`
	NATS       NATSConfig       `mapstructure:"nats"`
	Server     ServerConfig     `mapstructure:"server"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Auth       AuthConfig       `mapstructure:"auth"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type ClickHouseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type NATSConfig struct {
	URL string `mapstructure:"url"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type StorageConfig struct {
	Type   string `mapstructure:"type"` // s3, gcs, local
	Bucket string `mapstructure:"bucket"`
	Region string `mapstructure:"region"`
}

type AuthConfig struct {
	OpenFGAURL string `mapstructure:"openfga_url"`
	JWTSecret  string `mapstructure:"jwt_secret"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/agentflow")

	// Set defaults
	setDefaults()

	// Read environment variables
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Database defaults
	viper.SetDefault("database.host", getEnvOrDefault("DB_HOST", "localhost"))
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", getEnvOrDefault("DB_USER", "agentflow"))
	viper.SetDefault("database.password", getEnvOrDefault("DB_PASSWORD", ""))
	viper.SetDefault("database.database", getEnvOrDefault("DB_NAME", "agentflow"))
	viper.SetDefault("database.ssl_mode", "disable")

	// ClickHouse defaults
	viper.SetDefault("clickhouse.host", getEnvOrDefault("CLICKHOUSE_HOST", "localhost"))
	viper.SetDefault("clickhouse.port", 9000)
	viper.SetDefault("clickhouse.user", getEnvOrDefault("CLICKHOUSE_USER", "default"))
	viper.SetDefault("clickhouse.password", getEnvOrDefault("CLICKHOUSE_PASSWORD", ""))
	viper.SetDefault("clickhouse.database", getEnvOrDefault("CLICKHOUSE_DB", "agentflow"))

	// Redis defaults
	viper.SetDefault("redis.host", getEnvOrDefault("REDIS_HOST", "localhost"))
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", getEnvOrDefault("REDIS_PASSWORD", ""))
	viper.SetDefault("redis.db", 0)

	// NATS defaults
	viper.SetDefault("nats.url", getEnvOrDefault("NATS_URL", "nats://localhost:4222"))

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)

	// Storage defaults
	viper.SetDefault("storage.type", "local")
	viper.SetDefault("storage.bucket", "agentflow-artifacts")
	viper.SetDefault("storage.region", "us-east-1")

	// Auth defaults
	viper.SetDefault("auth.openfga_url", getEnvOrDefault("OPENFGA_URL", "http://localhost:8080"))
	viper.SetDefault("auth.jwt_secret", getEnvOrDefault("JWT_SECRET", "your-secret-key"))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
