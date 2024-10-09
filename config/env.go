package config

import "os"

func GetEnvVar(envVar string) string {
	return os.Getenv(envVar)
}

func DefaultEnvVar(envVar, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	return value
}
