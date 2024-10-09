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

	storageMgr, err := storage.NewStorageManager()
	if err != nil {
		log.Fatalf("failed to create storage manager: %v", err)
	}

	sessionManager := pkg.NewSessionManager()
	sessionManager.InitializeSessionCleaner()

	srv := server.Server{
		SessionMgr: sessionManager,
		StorageMgr: storageMgr,
	}

	r := gin.Default()
	SetupRoutes(srv, r)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func SetupRoutes(srv server.Server, r *gin.Engine) {

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(pkg.CORSMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	r.POST("/signup", srv.Signup)
	r.POST("/login", srv.Login)
	r.POST("/logout", srv.SessionMgr.Middleware, srv.Logout)
	r.POST("/delete", srv.SessionMgr.Middleware, srv.DeleteAccount)
	r.POST("/change-password", srv.SessionMgr.Middleware, srv.ChangePassword)

	r.POST("/credentials/store", srv.SessionMgr.Middleware, srv.StoreCreds)
	r.PUT("/credentials/edit", srv.SessionMgr.Middleware, srv.EditCreds)
	r.DELETE("/credentials/delete/:key", srv.SessionMgr.Middleware, srv.DeleteCreds)
	r.GET("/credentials/view/:key", srv.SessionMgr.Middleware, srv.ViewCreds)
	r.GET("/credentials/list", srv.SessionMgr.Middleware, srv.ListCreds)

	r.POST("/backup/:key", srv.SessionMgr.Middleware, srv.Backup)
	r.POST("/restore/:key", srv.SessionMgr.Middleware, srv.Restore)
	r.GET("/logs", srv.SessionMgr.Middleware, srv.Logs)
}
