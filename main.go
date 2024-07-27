package main

import (
	"dbhose/configs"
	"dbhose/handlers"
	"dbhose/utils"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
	configs.CheckEnvVars()
	configs.CheckPrograms()
}

func main() {

	r := gin.Default()
	SetupRoutes(r)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func SetupRoutes(r *gin.Engine) {

	sessionManager := utils.NewSessionManager()
	sessionManager.InitializeSessionCleaner()

	handler := handlers.Handler{
		SessionMgr: sessionManager,
	}

	r.POST("/signup", handler.Signup)
	r.POST("/login", handler.Login)
	r.POST("/logout", sessionManager.Middleware, handler.Logout)
	r.POST("/delete", sessionManager.Middleware, handler.DeleteAccount)
	r.POST("/change-password", sessionManager.Middleware, handler.ChangePassword)

	r.POST("/creds/store", sessionManager.Middleware, handlers.StoreCreds)
	r.PUT("/creds/edit", sessionManager.Middleware, handlers.EditCreds)
	r.DELETE("/creds/delete/:username", sessionManager.Middleware, handlers.DeleteCreds)
	r.GET("/creds/view/:username", sessionManager.Middleware, handlers.ViewCreds)

	r.POST("/backup", sessionManager.Middleware, handlers.Backup)
	r.POST("/restore", sessionManager.Middleware, handlers.Restore)
}
