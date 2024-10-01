package env

import (
	"os"
	"strconv"
	"time"
)

const (
	SECRET_KEY     = "SECRET_KEY"
	DEFAULT_SECRET = "secret"

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
