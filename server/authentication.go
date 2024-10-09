package server

import (
	"net/http"
	"time"

	"dbhose/domain"
	utils "dbhose/pkg"
	"dbhose/storage"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func (h *Server) Signup(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	if _, err := storage.GetUser(user.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	user.PasswordSalt = time.Now().Format(time.RFC3339Nano)
	saltedPassword := user.PasswordSalt + user.Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.Password = string(hashedPassword)

	// Store user in S3
	if err := storage.StoreUser(user); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": "Signup",
		}).Error("Failed to store user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := h.SessionMgr.CreateSession(user.Email)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": "Signup",
		}).Error("Failed to create session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful", "data": gin.H{"token": token}})
}

func (h *Server) Login(c *gin.Context) {
	var loginData domain.LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch user from S3
	user, err := storage.GetUser(loginData.Email)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to fetch user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or password"})
		return
	}

	// Compare passwords
	saltedPassword := user.PasswordSalt + loginData.Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(saltedPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or password"})
		return
	}

	// Generate JWT token
	token, err := h.SessionMgr.CreateSession(loginData.Email)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to create session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Server) Logout(c *gin.Context) {
	email := c.Value("email").(string)
	h.SessionMgr.DeleteSession(email)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (h *Server) DeleteAccount(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		utils.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to bind JSON")
		return
	}

	// Fetch user from S3
	storedUser, err := storage.GetUser(user.Email)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to fetch user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or password"})
		return
	}

	// Compare passwords
	saltedPassword := user.PasswordSalt + user.Password
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(saltedPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or password"})
		return
	}

	// Delete user from S3
	if err := storage.DeleteUser(user.Email); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := storage.DeleteAllCreds(user.Email); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to delete user credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.SessionMgr.DeleteSession(user.Email)

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func (h *Server) ChangePassword(c *gin.Context) {
	var changePasswordData domain.ChangePasswordData
	if err := c.ShouldBindJSON(&changePasswordData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch user from S3
	user, err := storage.GetUser(changePasswordData.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or password"})
		return
	}

	// Compare current password
	saltedPassword := user.PasswordSalt + changePasswordData.CurrentPassword
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(saltedPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid current password"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(changePasswordData.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update user password in S3
	user.Password = string(hashedPassword)
	if err := storage.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.SessionMgr.DeleteSession(user.Email)

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
