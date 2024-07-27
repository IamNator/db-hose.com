package main

import (
	"dbhose/configs"
	"dbhose/handlers"
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
	r.POST("/signup", handlers.Signup)
	r.POST("/login", handlers.Login)
	r.POST("/logout", handlers.Logout)
	r.POST("/delete", handlers.DeleteAccount)
	r.POST("/change-password", handlers.ChangePassword)

	r.POST("/backup", handlers.Backup)
	r.POST("/restore", handlers.Restore)
}
