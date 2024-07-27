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
	host := c.PostForm("host")
	port := c.PostForm("port")
	user := c.PostForm("user")
	dbname := c.PostForm("dbname")
	key := c.PostForm("key")

	// Fetch and decrypt credentials
	encryptedCreds, err := s3.GetEncryptedCredentials(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	password, err := utils.Decrypt(encryptedCreds, key)
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
	s3.LogBackup(user, key)

	utils.Log.WithFields(logrus.Fields{
		"event": "backup",
		"file":  key,
	}).Info("Backup successful")

	c.JSON(http.StatusOK, gin.H{"message": "Backup completed", "file": key})
}
