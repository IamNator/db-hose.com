package server

import (
	"dbhose/internal/storage"
	utils "dbhose/pkg"

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
func (h *Server) Run(port string) error {
	r := gin.Default()
	h.initRoutes(r)
	return r.Run(port)
}
