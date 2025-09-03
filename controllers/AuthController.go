package controllers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solchef/crypto-options-backend/config"
	"github.com/solchef/crypto-options-backend/models"
	"github.com/solchef/crypto-options-backend/utils"
	"gorm.io/gorm"
)

// --- helpers ---
func setRefreshCookie(c *gin.Context, token string, exp time.Time) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetCookie(
		"refresh_token",
		token,
		int(time.Until(exp).Seconds()),
		"/",
		"",     // domain
		secure, // secure in prod (HTTPS)
		true,   // httpOnly
	)
}

func clearRefreshCookie(c *gin.Context) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetCookie("refresh_token", "", -1, "/", "", secure, true)
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user and wallet
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.User true "User info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	var existing models.User
	if err := config.DB.Where("username = ?", input.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	input.Password = hashedPassword

	// Save to DB
	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	defaultCurrencies := []string{"USD", "BTC", "ETH"}
	for _, currency := range defaultCurrencies {
		wallet := models.Wallet{
			UserID:   input.ID,
			Currency: currency,
			Balance:  0,
		}
		config.DB.Create(&wallet)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login godoc
// @Summary Login a user
// @Description Authenticate and get JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body object{username=string,password=string} true "Login info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Print(input)

	var user models.User
	if err := config.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	// password hashing check you already have
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// access + refresh
	access, accessExp, err := utils.NewAccessToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "access token error"})
		return
	}
	refresh, jti, refreshExp, err := utils.NewRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token error"})
		return
	}

	// persist refresh
	rt := models.RefreshToken{
		UserID:    user.ID,
		JTI:       jti,
		ExpiresAt: refreshExp,
		Revoked:   false,
	}
	if err := config.DB.Create(&rt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist refresh"})
		return
	}

	// cookie (or you can also return in body)
	if os.Getenv("REFRESH_IN_COOKIE") != "false" {
		setRefreshCookie(c, refresh, refreshExp)
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": access,
		"expires_at":   accessExp.Unix(),
		// uncomment if you prefer returning also in body:
		"refresh_token": refresh,
	})
}

// Refresh godoc
// @Summary Refresh JWT tokens
// @Description Rotate refresh token and get a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body object{refresh_token=string} false "Refresh token in body (optional if in cookie)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security ApiKeyAuth
// @Router /auth/refresh [post]
func Refresh(c *gin.Context) {
	// try cookie first, fallback to body
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if berr := c.ShouldBindJSON(&body); berr != nil || body.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh token"})
			return
		}
		refreshToken = body.RefreshToken
	}

	claims, err := utils.ParseRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// find and ensure not revoked/expired
	var dbRT models.RefreshToken
	if err := config.DB.Where("jti = ? AND user_id = ?", claims.JTI, claims.Sub).First(&dbRT).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh not found"})
		return
	}
	if dbRT.Revoked || time.Now().After(dbRT.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh revoked or expired"})
		return
	}

	// rotate: revoke old, create new
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&dbRT).Update("revoked", true).Error; err != nil {
			return err
		}
		newRefresh, newJTI, newExp, err := utils.NewRefreshToken(claims.Sub)
		if err != nil {
			return err
		}
		if err := tx.Create(&models.RefreshToken{
			UserID:    claims.Sub,
			JTI:       newJTI,
			ExpiresAt: newExp,
			Revoked:   false,
		}).Error; err != nil {
			return err
		}

		// new access
		var user models.User
		if err := tx.First(&user, claims.Sub).Error; err != nil {
			return err
		}
		access, accessExp, err := utils.NewAccessToken(user.ID, user.Username)
		if err != nil {
			return err
		}

		// set cookie with new refresh
		if os.Getenv("REFRESH_IN_COOKIE") != "false" {
			setRefreshCookie(c, newRefresh, newExp)
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": access,
			"expires_at":   accessExp.Unix(),
			// "refresh_token": newRefresh, // if you also want in body
		})
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh failed"})
	}
}

// Logout godoc
// @Summary Logout user
// @Description Revoke refresh token and clear cookie
// @Tags auth
// @Produce json
// @Param refresh_token body object{refresh_token=string} false "Refresh token in body (optional if in cookie)"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	// revoke the refresh token presented (cookie/body)
	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.ShouldBindJSON(&body)
		refreshToken = body.RefreshToken
	}
	if refreshToken != "" {
		if claims, err := utils.ParseRefreshToken(refreshToken); err == nil {
			config.DB.Model(&models.RefreshToken{}).
				Where("jti = ? AND user_id = ?", claims.JTI, claims.Sub).
				Update("revoked", true)
		}
	}
	clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// Profile godoc
// @Summary Get user profile
// @Description Get authenticated user info
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /profile [get]
func Profile(c *gin.Context) {
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to your profile", "username": username})
}
