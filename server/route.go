package server

import (
	"dbhose/pkg"

	"github.com/gin-gonic/gin"
)

func (srv Server) initRoutes(r *gin.Engine) {
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(pkg.CORSMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	r.POST("/signup", srv.signup)
	r.POST("/login", srv.login)
	r.POST("/logout", srv.sessionMgr.Middleware, srv.logout)
	r.POST("/delete", srv.sessionMgr.Middleware, srv.deleteAccount)
	r.POST("/change-password", srv.sessionMgr.Middleware, srv.changePassword)

	r.POST("/credentials/store", srv.sessionMgr.Middleware, srv.storeCredential)
	r.PUT("/credentials/edit", srv.sessionMgr.Middleware, srv.editCredential)
	r.DELETE("/credentials/delete/:key", srv.sessionMgr.Middleware, srv.deleteCredential)
	r.GET("/credentials/view/:key", srv.sessionMgr.Middleware, srv.viewCredential)
	r.GET("/credentials/list", srv.sessionMgr.Middleware, srv.listCredential)

	r.POST("/backup/:key", srv.sessionMgr.Middleware, srv.backup)
	r.POST("/restore/:key", srv.sessionMgr.Middleware, srv.restore)
	r.GET("/logs", srv.sessionMgr.Middleware, srv.logs)
}
