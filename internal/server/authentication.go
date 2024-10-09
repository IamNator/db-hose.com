package server

import (
	"net/http"
	"time"

	"dbhose/internal/domain"
	"dbhose/internal/schema"
	"dbhose/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Sign up a new user
// @Description Create a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body schema.SignupRequest true "Signup Request"
// @Success 200 {object} schema.SignupResponse
// @Failure 400 {object} schema.ErrorResponse
// @Failure 409 {object} schema.ErrorResponse
// @Failure 422 {object} schema.ErrorResponse
// @Router /signup [post]
func (h *Server) signup(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	if _, err := h.storageMgr.GetUser(user.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	user.PasswordSalt = time.Now().Format(time.RFC3339Nano)
	saltedPassword := user.PasswordSalt + user.Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	user.Password = string(hashedPassword)

	// Store user in S3
	if err := h.storageMgr.StoreUser(user); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": "Signup",
		}).Error("Failed to store user")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	token, err := h.sessionMgr.CreateSession(user.Email)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": "Signup",
		}).Error("Failed to create session")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful", "data": gin.H{"token": token}})
}

// @Summary Log in a user
// @Description Authenticate a user and return a session token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body schema.LoginRequest true "Login Request"
// @Success 200 {object} schema.LoginResponse
// @Failure 400 {object} schema.ErrorResponse
// @Failure 401 {object} schema.ErrorResponse
// @Failure 422 {object} schema.ErrorResponse
// @Router /login [post]
func (h *Server) login(c *gin.Context) {
	var loginData schema.LoginRequest
	if err := c.ShouldBindJSON(&loginData); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch user from S3
	user, err := h.storageMgr.GetUser(loginData.Email)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to fetch user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email"})
		return
	}

	// Compare passwords
	saltedPassword := user.PasswordSalt + loginData.Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(saltedPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Password"})
		return
	}

	// Generate JWT token
	token, err := h.sessionMgr.CreateSession(loginData.Email)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to create session")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema.LoginResponse{Token: token})
}

// @Summary Log out a user
// @Description Invalidate the user's session
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} schema.GenericResponse
// @Failure 401 {object} schema.ErrorResponse
// @Router /logout [post]
func (h *Server) logout(c *gin.Context) {
	email := c.Value("email").(string)
	h.sessionMgr.DeleteSession(email)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// @Summary Delete a user account
// @Description Permanently delete a user account and associated data
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body schema.DeleteAccountRequest true "Account deletion request"
// @Security BearerAuth
// @Success 200 {object} schema.GenericResponse
// @Failure 400 {object} schema.ErrorResponse
// @Failure 401 {object} schema.ErrorResponse
// @Failure 422 {object} schema.ErrorResponse
// @Router /delete [post]
func (h *Server) deleteAccount(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		pkg.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to bind JSON")
		return
	}

	// Fetch user from S3
	storedUser, err := h.storageMgr.GetUser(user.Email)
	if err != nil {
		pkg.Log.WithFields(logrus.Fields{
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
	if err := h.storageMgr.DeleteUser(user.Email); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to delete user")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.storageMgr.DeleteAllCreds(user.Email); err != nil {
		pkg.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to delete user credentials")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	h.sessionMgr.DeleteSession(user.Email)

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

// @Summary Change user password
// @Description Change the password for a user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body schema.ChangePasswordRequest true "Password change request"
// @Security BearerAuth
// @Success 200 {object} schema.GenericResponse
// @Failure 400 {object} schema.ErrorResponse
// @Failure 401 {object} schema.ErrorResponse
// @Failure 422 {object} schema.ErrorResponse
// @Router /change-password [post]
func (h *Server) changePassword(c *gin.Context) {
	var changePasswordData schema.ChangePasswordRequest
	if err := c.ShouldBindJSON(&changePasswordData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch user from S3
	user, err := h.storageMgr.GetUser(changePasswordData.Email)
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
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Update user password in S3
	user.Password = string(hashedPassword)
	if err := h.storageMgr.UpdateUser(user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	h.sessionMgr.DeleteSession(user.Email)

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
