package server

import (
	utils "dbhose/pkg"
	"dbhose/storage"
)

type Server struct {
	SessionMgr *utils.SessionManager
	StorageMgr *storage.StorageManager
}
