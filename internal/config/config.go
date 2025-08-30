package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Elasticsearch ElasticsearchConfig
	Hold       HoldConfig
	RateLimit  RateLimitConfig
	Log        LogConfig
}

type AppConfig struct {
	Env  string
	Host string
	Port string
}

type DatabaseConfig struct {
	Host      string
	Port      string
	User      string
	Password  string
	Name      string
	Charset   string
	ParseTime bool
	Loc       string
}

type ElasticsearchConfig struct {
	Addresses []string
	Username  string
	Password  string
}

type HoldConfig struct {
	TTLMinutes int
	TTL        time.Duration
}

type RateLimitConfig struct {
	PerMinute int
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	holdTTLMinutes := getEnvAsInt("HOLD_TTL_MINUTES", 15)

	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Host: getEnv("APP_HOST", "0.0.0.0"),
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:      getEnv("DB_HOST", "localhost"),
			Port:      getEnv("DB_PORT", "3306"),
			User:      getEnv("DB_USER", "airline_user"),
			Password:  getEnv("DB_PASSWORD", "airline_pass"),
			Name:      getEnv("DB_NAME", "airline_booking"),
			Charset:   getEnv("DB_CHARSET", "utf8mb4"),
			ParseTime: getEnvAsBool("DB_PARSE_TIME", true),
			Loc:       getEnv("DB_LOC", "UTC"),
		},
		Elasticsearch: ElasticsearchConfig{
			Addresses: []string{getEnv("ES_ADDRESSES", "http://localhost:9200")},
			Username:  getEnv("ES_USERNAME", ""),
			Password:  getEnv("ES_PASSWORD", ""),
		},
		Hold: HoldConfig{
			TTLMinutes: holdTTLMinutes,
			TTL:        time.Duration(holdTTLMinutes) * time.Minute,
		},
		RateLimit: RateLimitConfig{
			PerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}, nil
}

func (d DatabaseConfig) DSN() string {
	return d.User + ":" + d.Password + "@tcp(" + d.Host + ":" + d.Port + ")/" + d.Name + "?charset=" + d.Charset + "&parseTime=" + strconv.FormatBool(d.ParseTime) + "&loc=" + d.Loc
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
