package handlers

import (
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

func Restore(c *gin.Context) {

	email := c.PostForm("email")
	key := c.PostForm("key")
	secret := c.Query("secret")
	fileName := c.Query("file")

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

	// Download the file from S3
	result, err := s3.DownloadFromS3(bucket, fileName)
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

	// Write the S3 object to the command input
	if _, err := io.Copy(inPipe, result.Body); err != nil {
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
	s3.LogRestore(duration, email, fileName)

	utils.Log.WithFields(logrus.Fields{
		"event": "restore",
		"file":  fileName,
	}).Info("Restore successful")

	c.JSON(http.StatusOK, gin.H{"message": "Restore completed"})
}
