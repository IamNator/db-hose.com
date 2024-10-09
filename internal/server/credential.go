package server

import (
	"dbhose/internal/domain"
	"dbhose/pkg"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// storeCredential stores encrypted credentials in S3
func (h *Server) storeCredential(c *gin.Context) {

	email := c.Value("email").(string)
	secret := c.Query("secret")

	var creds domain.Credential
	if err := c.BindJSON(&creds); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "storeCreds",
			"error": err.Error(),
		}).Error("Invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	for key, value := range creds.Secret {

		encryptedValue, err := pkg.Encrypt(value, secret)
		if err != nil {
			pkg.Log.WithFields(logrus.Fields{
				"event": "storeCreds",
				"error": err.Error(),
			}).Error("Encryption failed")
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Encryption failed"})
			return
		}
		creds.Secret[key] = encryptedValue
	}

	if err := h.storageMgr.StoreCreds(email, creds); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "storeCreds",
			"error": err.Error(),
		}).Error("Failed to store credentials")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to store credentials"})
		return
	}

	pkg.Log.WithFields(logrus.Fields{
		"event": "storeCreds",
		"creds": creds,
	}).Info("Credentials stored successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Credentials stored successfully"})
}

// editCredential edits stored credentials in S3
func (h *Server) editCredential(c *gin.Context) {
	email := c.Value("email").(string)
	secret := c.Query("secret")

	var creds domain.Credential
	if err := c.BindJSON(&creds); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "editCreds",
			"error": err.Error(),
		}).Error("Invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	savedCreds, err := h.storageMgr.FindCredentialByID(email, creds.ID)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "editCreds",
			"error": err.Error(),
		}).Error("Failed to retrieve credentials")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to retrieve credentials"})
		return
	}

	for key, value := range creds.Secret {

		if savedValue, ok := savedCreds.Secret[key]; ok {
			if savedValue == value {
				continue
			}
		}

		encryptedValue, err := pkg.Encrypt(value, secret)
		if err != nil {
			pkg.Log.WithFields(logrus.Fields{
				"event": "editCreds",
				"error": err.Error(),
			}).Error("Encryption failed")
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Encryption failed"})
			return
		}
		creds.Secret[key] = encryptedValue
	}

	if err := h.storageMgr.UpdateCreds(email, creds); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "editCreds",
			"error": err.Error(),
		}).Error("Failed to edit credentials")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to edit credentials"})
		return
	}

	pkg.Log.WithFields(logrus.Fields{
		"event": "editCreds",
		"creds": creds,
	}).Info("Credentials edited successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Credentials edited successfully"})
}

// deleteCredential deletes stored credentials from S3
func (h *Server) deleteCredential(c *gin.Context) {

	email := c.Value("email").(string)
	key := c.Param("key")

	if err := h.storageMgr.DeleteCreds(email, key); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "deleteCreds",
			"error": err.Error(),
		}).Error("Failed to delete credentials")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to delete credentials"})
		return
	}

	pkg.Log.WithFields(logrus.Fields{
		"event": "deleteCreds",
		"key":   key,
	}).Info("Credentials deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Credentials deleted successfully"})
}

// viewCredential views stored credentials from S3
func (h *Server) viewCredential(c *gin.Context) {

	email := c.Value("email").(string)
	secret := c.Query("secret")
	key := c.Param("key")

	creds, err := h.storageMgr.FindCredentialByID(email, key)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "viewCreds",
			"key":   key,
			"error": err.Error(),
		}).Error("Failed to retrieve credentials")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to retrieve credentials"})
		return
	}

	if secret != "" {
		for k, encryptedValue := range creds.Secret {
			decryptedValue, err := pkg.Decrypt(encryptedValue, secret)
			if err != nil {
				pkg.Log.WithFields(logrus.Fields{
					"event": "viewCreds",
					"error": err.Error(),
				}).Error("Decryption failed")
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Decryption failed"})
				return
			}
			creds.Secret[k] = decryptedValue
		}
	}

	pkg.Log.WithFields(logrus.Fields{
		"event": "viewCreds",
		"creds": creds,
	}).Info("Credentials retrieved successfully")

	c.JSON(http.StatusOK, gin.H{"credentials": creds})
}

// listCredential lists stored credentials from S3
func (h *Server) listCredential(c *gin.Context) {

	email := c.Value("email").(string)
	creds, err := h.storageMgr.ListCredential(email)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"event": "listCreds",
			"error": err.Error(),
		}).Error("Failed to list credentials")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to list credentials"})
		return
	}

	pkg.Log.WithFields(logrus.Fields{
		"event": "listCreds",
		"creds": creds,
	}).Info("Credentials listed successfully")

	c.JSON(http.StatusOK, gin.H{"credentials": creds})
}
