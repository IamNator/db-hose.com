package main

import (
	"dbhose/config"
	utils "dbhose/pkg"
	handlers "dbhose/server"
	s3 "dbhose/storage"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
	config.CheckEnvVars()
	config.CheckPrograms()
	s3.Init()
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

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(utils.CORSMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	r.POST("/signup", handler.Signup)
	r.POST("/login", handler.Login)
	r.POST("/logout", sessionManager.Middleware, handler.Logout)
	r.POST("/delete", sessionManager.Middleware, handler.DeleteAccount)
	r.POST("/change-password", sessionManager.Middleware, handler.ChangePassword)

	r.POST("/creds/store", sessionManager.Middleware, handlers.StoreCreds)
	r.PUT("/creds/edit", sessionManager.Middleware, handlers.EditCreds)
	r.DELETE("/creds/delete/:key", sessionManager.Middleware, handlers.DeleteCreds)
	r.GET("/creds/view/:key", sessionManager.Middleware, handlers.ViewCreds)
	r.GET("/creds/list", sessionManager.Middleware, handlers.ListCreds)

	r.POST("/backup/:key", sessionManager.Middleware, handlers.Backup)
	r.POST("/restore/:key", sessionManager.Middleware, handlers.Restore)
	r.GET("/logs", sessionManager.Middleware, handlers.Logs)
}
