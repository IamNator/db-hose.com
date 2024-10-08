package server

import (
	"bytes"
	utils "dbhose/pkg"
	"dbhose/storage"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Backup(c *gin.Context) {

	email := c.Value("email").(string)
	key := c.Param("key")

	secret := c.Query("secret")

	// Fetch and decrypt credentials
	encryptedCreds, err := storage.GetCreds(email, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	encryptedUser := encryptedCreds.Values["user"]
	encryptedPassword := encryptedCreds.Values["password"]
	encryptedHost := encryptedCreds.Values["host"]
	encryptedPort := encryptedCreds.Values["port"]
	encryptedDBName := encryptedCreds.Values["dbname"]

	user, err := utils.Decrypt(encryptedUser, secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	password, err := utils.Decrypt(encryptedPassword, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	host, err := utils.Decrypt(encryptedHost, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	port, err := utils.Decrypt(encryptedPort, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dbname, err := utils.Decrypt(encryptedDBName, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create the command
	cmdStr := fmt.Sprintf(`pg_dump --column-inserts --no-owner "host=%s port=%s user=%s dbname=%s password=%s sslmode=require"`, host, port, user, dbname, password)
	cmd := exec.Command("sh", "-c", cmdStr)

	// Create a pipe for the output
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()

	// Start the command
	if err := cmd.Start(); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "backup",
			"error": err.Error(),
		}).Error("Backup failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Backup failed", "details": err.Error()})
		return
	}

	duration := time.Since(start)

	// Upload the output to S3
	fileName := fmt.Sprintf("%s.psql.gz", time.Now().Format("2006-01-02"))
	key = "backups/" + fileName
	body := bytes.NewReader(buf.Bytes())

	if err := storage.StoreBackup(key, body); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "backup",
			"error": err.Error(),
		}).Error("Failed to upload backup to S3")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload backup", "details": err.Error()})
		return
	}

	// Log the backup
	storage.LogBackup(duration, user, key)

	utils.Log.WithFields(logrus.Fields{
		"event": "backup",
		"file":  key,
	}).Info("Backup successful")

	c.JSON(http.StatusOK, gin.H{"message": "Backup completed", "file": key})
}

func Restore(c *gin.Context) {

	key := c.Param("key")
	email := c.Value("email").(string)

	secret := c.Query("secret")
	fileName := c.Query("file")

	// Fetch and decrypt credentials
	encryptedCreds, err := storage.GetCreds(email, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	encryptedUser := encryptedCreds.Values["user"]
	encryptedPassword := encryptedCreds.Values["password"]
	encryptedHost := encryptedCreds.Values["host"]
	encryptedPort := encryptedCreds.Values["port"]
	encryptedDBName := encryptedCreds.Values["dbname"]

	user, err := utils.Decrypt(encryptedUser, secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	password, err := utils.Decrypt(encryptedPassword, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	host, err := utils.Decrypt(encryptedHost, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	port, err := utils.Decrypt(encryptedPort, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dbname, err := utils.Decrypt(encryptedDBName, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Download the file from S3
	result, err := storage.FetchBackup(fileName)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "restore",
			"error": err.Error(),
		}).Error("Failed to download backup from S3")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download backup", "details": err.Error()})
		return
	}

	// Create the command
	cmdStr := fmt.Sprintf(`psql "host=%s port=%s user=%s dbname=%s password=%s sslmode=require"`, host, port, user, dbname, password)
	cmd := exec.Command("sh", "-c", cmdStr)

	// Create a pipe for the input
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()

	// Start the command
	if err := cmd.Start(); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "restore",
			"error": err.Error(),
		}).Error("Restore failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Restore failed", "details": err.Error()})
		return
	}

	bodyReader := bytes.NewReader(result)

	// Write the S3 object to the command input
	if _, err := io.Copy(inPipe, bodyReader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inPipe.Close()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "restore",
			"error": err.Error(),
		}).Error("Restore failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Restore failed", "details": err.Error()})
		return
	}

	duration := time.Since(start)

	// Log the restore
	storage.LogRestore(duration, email, fileName)

	utils.Log.WithFields(logrus.Fields{
		"event": "restore",
		"file":  fileName,
	}).Info("Restore successful")

	c.JSON(http.StatusOK, gin.H{"message": "Restore completed"})
}

func Logs(c *gin.Context) {
	// Logs
	email := c.Value("email").(string)

	logs, err := storage.FetchLogs(email)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "logs",
			"error": err.Error(),
		}).Error("Failed to fetch logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logs fetched successfully", "data": logs})
}
