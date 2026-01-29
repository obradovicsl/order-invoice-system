package loader

import (
	"fmt"
	"os"
	"strconv"
)

func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse environment variable %s=%q as integer: %v", key, valueStr, err))
		}
		return value
	}
	return defaultValue
}

func GetEnvOrCrash(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return value
}

func GetEnvAsBool(key string, defaultValue bool) bool {
	valueStr, isPresent := os.LookupEnv(key)

	if !isPresent {
		return defaultValue
	}

	if valueStr != "true" && valueStr != "false" {
		panic(fmt.Sprintf("Failed to parse environment variable %s=%q as boolean", key, valueStr))
	}

	return valueStr == "true"
}
