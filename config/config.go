package config

import (
	"log"
	"os"
	"os/exec"
)

var requiredEnvVars = []string{
	"AWS_REGION",
	"S3_BUCKET_NAME",
	"JWT_SECRET_KEY",
}

var requiredPrograms = []string{
	"pg_dump",
	"psql",
	"gzip",
	"gunzip",
}

func CheckEnvVars() {
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Environment variable %s is not set", envVar)
		}
	}
}

func CheckPrograms() {
	for _, program := range requiredPrograms {
		_, err := exec.LookPath(program)
		if err != nil {
			log.Fatalf("Program %s is not installed or not found in PATH", program)
		}
	}
}
