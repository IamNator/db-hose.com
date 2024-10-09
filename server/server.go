package server

import (
	"dbhose/config"
	utils "dbhose/pkg"
	"dbhose/storage"

	"github.com/gin-gonic/gin"
)

type Server struct {
	sessionMgr *utils.SessionManager
	storageMgr *storage.StorageManager
}

// New creates a new server instance
func New(sessionMgr *utils.SessionManager, storageMgr *storage.StorageManager) *Server {
	return &Server{
		sessionMgr: sessionMgr,
		storageMgr: storageMgr,
	}
}

// Run starts the server
func (h *Server) Run(r *gin.Engine) error {
	h.initRoutes(r)
	port := config.DefaultEnvVar("PORT", ":8080")
	return r.Run(port)
}
