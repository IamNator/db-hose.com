package handlers

import (
	"bytes"
	"dbhose/s3"
	"dbhose/utils"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	bucket = "migrations"
)

func Backup(c *gin.Context) {

	email := c.PostForm("email")
	key := c.PostForm("key")
	secret := c.Query("secret")

	// Fetch and decrypt credentials
	encryptedCreds, err := s3.GetCreds(email, key)
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

	if err := s3.UploadToS3(bucket, key, body); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "backup",
			"error": err.Error(),
		}).Error("Failed to upload backup to S3")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload backup", "details": err.Error()})
		return
	}

	// Log the backup
	s3.LogBackup(duration, user, key)

	utils.Log.WithFields(logrus.Fields{
		"event": "backup",
		"file":  key,
	}).Info("Backup successful")

	c.JSON(http.StatusOK, gin.H{"message": "Backup completed", "file": key})
}
