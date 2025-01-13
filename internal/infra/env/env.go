package env

import (
	"os"
	"strconv"
	"time"
)

const (
	SERVICE_NAME         = "SERVICE_NAME"
	VERSION              = "SERVICE_VERSION"
	DEFAULT_SERVICE_NAME = "auth"
	DEFAULT_VERSION      = "v1.0.0"

	LOG_LEVEL   = "LOG_LEVEL"
	DEFAULT_LOG = "info"

	ENVIRONMENT  = "ENVIRONMENT"
	DEFAULT_ENV  = "local"
	PORT         = "PORT"
	DEFAULT_PORT = "8080"

	SECRET_KEY     = "SECRET_KEY"
	DEFAULT_SECRET = "secret"

	HTTP_TIMEOUT         = "HTTP_TIMEOUT"
	DEFAULT_HTTP_TIMEOUT = 30 * time.Second

	OTEL_EXPORTER_OTLP_ENDPOINT         = "OTEL_EXPORTER_OTLP_ENDPOINT"
	DEFAULT_OTEL_EXPORTER_OTLP_ENDPOINT = "localhost:4318"

	MYSQL_USER             = "MYSQL_USER"
	MYSQL_PASS             = "MYSQL_PASS"
	MYSQL_HOST             = "MYSQL_HOST"
	MYSQL_DATABASE         = "MYSQL_DATABASE"
	DEFAULT_MYSQL_USER     = "root"
	DEFAULT_MYSQL_PASS     = "auth"
	DEFAULT_MYSQL_HOST     = "localhost:3306"
	DEFAULT_MYSQL_DATABASE = "auth"

	REDIS_HOST         = "REDIS_HOST"
	DEFAULT_REDIS_HOST = "localhost:6379"

	ACCESS_TOKEN_EXPIRATION_SECONDS  = "ACCESS_TOKE_EXPIRATION_SECONDS"
	REFRESH_TOKEN_EXPIRATION         = "REFRESH_TOKEN_EXPIRATION_SECONDS"
	DEFAULT_ACCESS_TOKEN_EXPIRATION  = 5 * time.Minute
	DEFAULT_REFRESH_TOKEN_EXPIRATION = 1 * time.Hour
)

func GetEnvAsString(value string, defaultValue string) string {
	result, found := os.LookupEnv(value)
	if found {
		return result
	}
	return defaultValue
}

func GetEnvAsInt64(value string, defaultValue int64) int64 {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return result
}

func GetEnvAsFloat64(value string, defaultValue float64) float64 {
	result, err := strconv.ParseFloat(value, 10)
	if err != nil {
		return defaultValue
	}
	return result
}
