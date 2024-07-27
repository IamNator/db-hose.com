package handlers

import (
	"dbhose/s3"
	"dbhose/utils"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Restore(c *gin.Context) {

	
	host := c.PostForm("host")
	port := c.PostForm("port")
	dbname := c.PostForm("dbname")
	email := c.PostForm("email")
	fileKey := c.PostForm("file")
	key := c.PostForm("key")

	// Fetch and decrypt credentials
	encryptedCreds, err := s3.GetEncryptedCredentials(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	password, err := utils.Decrypt(encryptedCreds, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Download the file from S3
	result, err := s3.DownloadFromS3(bucket, fileKey)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "restore",
			"error": err.Error(),
		}).Error("Failed to download backup from S3")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download backup", "details": err.Error()})
		return
	}

	defer result.Body.Close()

	// Create the command
	cmdStr := fmt.Sprintf(`psql "host=%s port=%s user=%s dbname=%s password=%s sslmode=require"`, host, port, email, dbname, password)
	cmd := exec.Command("sh", "-c", cmdStr)

	// Create a pipe for the input
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "restore",
			"error": err.Error(),
		}).Error("Restore failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Restore failed", "details": err.Error()})
		return
	}

	// Write the S3 object to the command input
	if _, err := io.Copy(inPipe, result.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inPipe.Close()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log the restore
	s3.LogRestore(email, fileKey)

	utils.Log.WithFields(logrus.Fields{
		"event": "restore",
		"file":  fileKey,
	}).Info("Restore successful")

	c.JSON(http.StatusOK, gin.H{"message": "Restore completed"})
}
