package server

import (
	"dbhose/pkg"
	"net/http"
	"os"
	"path/filepath"

	_ "dbhose/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (srv Server) initRoutes(engine *gin.Engine) {

	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())
	engine.Use(pkg.CORSMiddleware())

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	engine.Static("/public", "./public")

	engine.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	engine.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./public/favicon.ico")
	})

	// Handle 404 errors
	engine.NoRoute(func(c *gin.Context) {
		path := filepath.Join("public", c.Request.URL.Path)

		if info, err := os.Stat(path); err == nil && info.IsDir() {
			path = filepath.Join(path, "index.html")
		}

		if _, err := os.ReadFile(path); err == nil {
			c.File(path)
		} else {
			c.String(http.StatusNotFound, "404 page not found")
		}
	})

	eng := engine.Group("/api/v1")

	eng.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	eng.POST("/signup", srv.signup)
	eng.POST("/login", srv.login)
	eng.POST("/logout", srv.sessionMgr.Middleware, srv.logout)
	eng.POST("/delete", srv.sessionMgr.Middleware, srv.deleteAccount)
	eng.POST("/change-password", srv.sessionMgr.Middleware, srv.changePassword)

	eng.POST("/credentials/store", srv.sessionMgr.Middleware, srv.storeCredential)
	eng.PUT("/credentials/edit", srv.sessionMgr.Middleware, srv.editCredential)
	eng.DELETE("/credentials/delete/:key", srv.sessionMgr.Middleware, srv.deleteCredential)
	eng.GET("/credentials/view/:key", srv.sessionMgr.Middleware, srv.viewCredential)
	eng.GET("/credentials/list", srv.sessionMgr.Middleware, srv.listCredential)

	eng.POST("/backup/:key", srv.sessionMgr.Middleware, srv.backup)
	eng.POST("/restore/:key", srv.sessionMgr.Middleware, srv.restore)
	eng.GET("/logs", srv.sessionMgr.Middleware, srv.logs)
	eng.GET("/migration", srv.sessionMgr.Middleware, srv.migrationHistory)
}
