package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"dbhose/config"
	"dbhose/pkg"
	"dbhose/server"
	"dbhose/storage"
)

func init() {
	godotenv.Load()
	config.CheckEnvVars()
	config.CheckPrograms()
}

func main() {

	storageMgr, err := storage.New()
	if err != nil {
		log.Fatalf("failed to create storage manager: %v", err)
	}

	sessionManager := pkg.NewSessionManager()
	sessionManager.InitializeSessionCleaner()

	srv := server.New(sessionManager, storageMgr)

	r := gin.Default()
	if err := srv.Run(r); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
