package server

import (
	models "dbhose/domain"
	utils "dbhose/pkg"
	s3 "dbhose/storage"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// StoreCreds stores encrypted credentials in S3
func StoreCreds(c *gin.Context) {

	email := c.Value("email").(string)
	secret := c.Query("secret")

	var creds models.Credentials
	if err := c.BindJSON(&creds); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "storeCreds",
			"error": err.Error(),
		}).Error("Invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	for key, value := range creds.Values {

		encryptedValue, err := utils.Encrypt(value, secret)
		if err != nil {
			utils.Log.WithFields(logrus.Fields{
				"event": "storeCreds",
				"error": err.Error(),
			}).Error("Encryption failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
			return
		}
		creds.Values[key] = encryptedValue
	}

	if err := s3.StoreCreds(email, creds); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "storeCreds",
			"error": err.Error(),
		}).Error("Failed to store credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store credentials"})
		return
	}

	utils.Log.WithFields(logrus.Fields{
		"event": "storeCreds",
		"creds": creds,
	}).Info("Credentials stored successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Credentials stored successfully"})
}

// EditCreds edits stored credentials in S3
func EditCreds(c *gin.Context) {
	email := c.Value("email").(string)
	secret := c.Query("secret")

	var creds models.Credentials
	if err := c.BindJSON(&creds); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "editCreds",
			"error": err.Error(),
		}).Error("Invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	savedCreds, err := s3.GetCreds(email, creds.Key)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "editCreds",
			"error": err.Error(),
		}).Error("Failed to retrieve credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve credentials"})
		return
	}

	for key, value := range creds.Values {

		if savedValue, ok := savedCreds.Values[key]; ok {
			if savedValue == value {
				continue
			}
		}

		encryptedValue, err := utils.Encrypt(value, secret)
		if err != nil {
			utils.Log.WithFields(logrus.Fields{
				"event": "editCreds",
				"error": err.Error(),
			}).Error("Encryption failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
			return
		}
		creds.Values[key] = encryptedValue
	}

	if err := s3.UpdateCreds(email, creds); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "editCreds",
			"error": err.Error(),
		}).Error("Failed to edit credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit credentials"})
		return
	}

	utils.Log.WithFields(logrus.Fields{
		"event": "editCreds",
		"creds": creds,
	}).Info("Credentials edited successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Credentials edited successfully"})
}

// DeleteCreds deletes stored credentials from S3
func DeleteCreds(c *gin.Context) {

	email := c.Value("email").(string)
	key := c.Param("key")

	if err := s3.DeleteCreds(email, key); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "deleteCreds",
			"error": err.Error(),
		}).Error("Failed to delete credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete credentials"})
		return
	}

	utils.Log.WithFields(logrus.Fields{
		"event": "deleteCreds",
		"key":   key,
	}).Info("Credentials deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Credentials deleted successfully"})
}

// ViewCreds views stored credentials from S3
func ViewCreds(c *gin.Context) {

	email := c.Value("email").(string)
	secret := c.Query("secret")
	key := c.Param("key")

	creds, err := s3.GetCreds(email, key)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "viewCreds",
			"key":   key,
			"error": err.Error(),
		}).Error("Failed to retrieve credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve credentials"})
		return
	}

	if secret != "" {
		for k, encryptedValue := range creds.Values {
			decryptedValue, err := utils.Decrypt(encryptedValue, secret)
			if err != nil {
				utils.Log.WithFields(logrus.Fields{
					"event": "viewCreds",
					"error": err.Error(),
				}).Error("Decryption failed")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Decryption failed"})
				return
			}
			creds.Values[k] = decryptedValue
		}
	}

	utils.Log.WithFields(logrus.Fields{
		"event": "viewCreds",
		"creds": creds,
	}).Info("Credentials retrieved successfully")

	c.JSON(http.StatusOK, gin.H{"credentials": creds})
}

// ListCreds lists stored credentials from S3
func ListCreds(c *gin.Context) {

	email := c.Value("email").(string)
	creds, err := s3.ListCreds(email)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"event": "listCreds",
			"error": err.Error(),
		}).Error("Failed to list credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list credentials"})
		return
	}

	utils.Log.WithFields(logrus.Fields{
		"event": "listCreds",
		"creds": creds,
	}).Info("Credentials listed successfully")

	c.JSON(http.StatusOK, gin.H{"credentials": creds})
}
