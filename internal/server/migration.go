package server

import (
	"bytes"
	"dbhose/internal/domain"
	"dbhose/internal/schema"
	"dbhose/pkg"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @Summary Backup a database
// @Description Backup a database
// @Tags Migration
// @Accept json
// @Produce json
// @Param key path string true "Credential key"
// @Param secret query string true "request"
// @Security Bearer
// @Success 200 {object} schema.GenericResponse
// @Failure 400 {object} schema.ErrorResponse
// @Failure 422 {object} schema.ErrorResponse
// @Router /backup/{key} [post]
func (h *Server) backup(c *gin.Context) {
	email := c.Value("email").(string)
	key := c.Param("key")

	secret := c.Query("secret")

	// Fetch and decrypt credentials
	encryptedCreds, err := h.storageMgr.FindCredentialByID(email, key)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := encryptedCreds.Decrypt(secret); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Create the command
	cmdStr := fmt.Sprintf(`pg_dump --column-inserts --no-owner "host=%s port=%s user=%s dbname=%s password=%s sslmode=require"`,
		encryptedCreds.Secret.Host, encryptedCreds.Secret.Port, encryptedCreds.Secret.User, encryptedCreds.Secret.DBName, encryptedCreds.Secret.Password)
	cmd := exec.Command("sh", "-c", cmdStr)

	// Create a pipe for the output
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()

	// Start the command
	if err := cmd.Start(); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "backup",
			"error": err.Error(),
		}).Error("Backup failed")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Read the command output
	var buf bytes.Buffer
	writer := io.MultiWriter(&buf)
	go func() {
		io.Copy(writer, outPipe)
	}()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Backup failed", "details": err.Error()})
		return
	}

	duration := time.Since(start)

	// Upload the output to S3
	fileName := fmt.Sprintf("%s.psql.gz", time.Now().Format("2006-01-02"))
	key = "backups/" + fileName
	body := bytes.NewReader(buf.Bytes())

	if err := h.storageMgr.StoreBackup(key, body); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "backup",
			"error": err.Error(),
		}).Error("Failed to upload backup to S3")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to upload backup", "details": err.Error()})
		return
	}

	// Log the backup
	h.storageMgr.LogBackup(duration, encryptedCreds.Secret.User, key)

	pkg.Log.WithFields(logrus.Fields{
		"event": "backup",
		"file":  key,
	}).Info("Backup successful")

	c.JSON(http.StatusOK, gin.H{"message": "Backup completed", "file": key})
}

// @Summary Restore a database
// @Description Restore a database
// @Tags Migration
// @Accept json
// @Produce json
// @Param key path string true "Credential key"
// @Param secret query string true "request"
// @Param file query string true "Backup file"
// @Security Bearer
// @Success 200 {object} schema.GenericResponse
// @Failure 400 {object} schema.ErrorResponse
// @Failure 422 {object} schema.ErrorResponse
// @Router /restore/{key} [post]
func (h *Server) restore(c *gin.Context) {

	key := c.Param("key")
	email := c.Value("email").(string)

	secret := c.Query("secret")
	fileName := c.Query("file")

	// Fetch and decrypt credentials
	encryptedCreds, err := h.storageMgr.FindCredentialByID(email, key)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := encryptedCreds.Decrypt(secret); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Download the file from S3
	result, err := h.storageMgr.FetchBackup(fileName, time.Now())
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "restore",
			"error": err.Error(),
		}).Error("Failed to download backup from S3")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to download backup", "details": err.Error()})
		return
	}

	// Create the command
	cmdStr := fmt.Sprintf(`psql "host=%s port=%s user=%s dbname=%s password=%s sslmode=require"`,
		encryptedCreds.Secret.Host, encryptedCreds.Secret.Port, encryptedCreds.Secret.User, encryptedCreds.Secret.DBName, encryptedCreds.Secret.Password)
	cmd := exec.Command("sh", "-c", cmdStr)

	// Create a pipe for the input
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()

	// Start the command
	if err := cmd.Start(); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "restore",
			"error": err.Error(),
		}).Error("Restore failed")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Restore failed", "details": err.Error()})
		return
	}

	bodyReader := bytes.NewReader(result)

	// Write the S3 object to the command input
	if _, err := io.Copy(inPipe, bodyReader); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	inPipe.Close()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "restore",
			"error": err.Error(),
		}).Error("Restore failed")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Restore failed", "details": err.Error()})
		return
	}

	duration := time.Since(start)

	// Log the restore
	h.storageMgr.LogRestore(duration, email, fileName)

	pkg.Log.WithFields(logrus.Fields{
		"event": "restore",
		"file":  fileName,
	}).Info("Restore successful")

	c.JSON(http.StatusOK, gin.H{"message": "Restore completed"})
}

// @Summary Fetch logs
// @Description Fetch logs
// @Tags Migration
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} schema.LogResponse
// @Failure 400 {object} schema.ErrorResponse
// @Router /logs [get]
func (h *Server) logs(c *gin.Context) {
	email := c.Value("email").(string)

	logs, err := h.storageMgr.FetchLogs(email)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "logs",
			"error": err.Error(),
		}).Error("Failed to fetch logs")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to fetch logs", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logs fetched successfully", "data": logs})
}

type migrationHistoryResponse schema.Response[[]domain.Migration]

// @Summary Fetch Migration History
// @Description Fetch Migration History
// @Tags Migration
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} migrationHistoryResponse
// @Failure 400 {object} schema.ErrorResponse
// @Router /migration [get]
func (h *Server) migrationHistory(c *gin.Context) {
	email := c.Value("email").(string)

	migrationHistory, err := h.storageMgr.ListBackups(email)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "migration-history",
			"error": err.Error(),
		}).Error("Failed to fetch migration history")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to fetch migration history", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, migrationHistoryResponse{Message: "Migration history fetched successfully", Data: migrationHistory})
}
