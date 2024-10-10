package main

import (
	"log"

	"github.com/joho/godotenv"

	"dbhose/config"
	"dbhose/internal/server"
	"dbhose/internal/storage"
	"dbhose/pkg"
)

func init() {
	godotenv.Load()
	config.CheckEnvVars()
	config.CheckPrograms()
}

// @title DBHose API
// @version 1.0
// @description This is the API for DBHose
// @host localhost:8080
// @BasePath /api/v1
func main() {
	storageMgr, err := storage.New()
	if err != nil {
		log.Fatalf("failed to create storage manager: %v", err)
	}

	sessionManager := pkg.NewSessionManager()
	sessionManager.InitializeSessionCleaner()

	srv := server.New(sessionManager, storageMgr)

	port := config.DefaultEnvVar("PORT", ":8080")

	if err := srv.Run(port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
