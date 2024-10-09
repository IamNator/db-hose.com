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
	storage.Init()
}

func main() {
	r := gin.Default()
	SetupRoutes(r)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func SetupRoutes(r *gin.Engine) {

	sessionManager := pkg.NewSessionManager()
	sessionManager.InitializeSessionCleaner()

	handler := server.Server{
		SessionMgr: sessionManager,
	}

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(pkg.CORSMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	r.POST("/signup", handler.Signup)
	r.POST("/login", handler.Login)
	r.POST("/logout", sessionManager.Middleware, handler.Logout)
	r.POST("/delete", sessionManager.Middleware, handler.DeleteAccount)
	r.POST("/change-password", sessionManager.Middleware, handler.ChangePassword)

	r.POST("/credentials/store", sessionManager.Middleware, server.StoreCreds)
	r.PUT("/credentials/edit", sessionManager.Middleware, server.EditCreds)
	r.DELETE("/credentials/delete/:key", sessionManager.Middleware, server.DeleteCreds)
	r.GET("/credentials/view/:key", sessionManager.Middleware, server.ViewCreds)
	r.GET("/credentials/list", sessionManager.Middleware, server.ListCreds)

	r.POST("/backup/:key", sessionManager.Middleware, server.Backup)
	r.POST("/restore/:key", sessionManager.Middleware, server.Restore)
	r.GET("/logs", sessionManager.Middleware, server.Logs)
}
