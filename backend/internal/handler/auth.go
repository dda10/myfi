package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"myfi-backend/internal/model"
)

// HandleRegister handles POST /api/auth/register
func (h *Handlers) HandleRegister(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.AuthService.Register(req)
	if err != nil {
		if authErr, ok := err.(*model.AuthError); ok {
			if authErr.Code == model.ErrCodeUserExists {
				c.JSON(http.StatusConflict, gin.H{"error": authErr.Message, "code": authErr.Code})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// HandleLogin handles POST /api/auth/login
func (h *Handlers) HandleLogin(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.AuthService.Login(req)
	if err != nil {
		if authErr, ok := err.(*model.AuthError); ok {
			switch authErr.Code {
			case model.ErrCodeInvalidCredentials:
				c.JSON(http.StatusUnauthorized, gin.H{"error": authErr.Message, "code": authErr.Code})
			case model.ErrCodeAccountLocked:
				c.JSON(http.StatusTooManyRequests, gin.H{"error": authErr.Message, "code": authErr.Code})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{"error": authErr.Message, "code": authErr.Code})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleLogout handles POST /api/auth/logout
func (h *Handlers) HandleLogout(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	jwtClaims := claims.(*model.JWTClaims)
	_ = h.AuthService.Logout(jwtClaims.SessionID)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// HandleChangePassword handles POST /api/auth/change-password
func (h *Handlers) HandleChangePassword(c *gin.Context) {
	claims := c.MustGet("claims").(*model.JWTClaims)

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.AuthService.ChangePassword(claims.UserID, req); err != nil {
		if authErr, ok := err.(*model.AuthError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": authErr.Message, "code": authErr.Code})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password change failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// HandleMe handles GET /api/auth/me
func (h *Handlers) HandleMe(c *gin.Context) {
	claims := c.MustGet("claims").(*model.JWTClaims)

	user, err := h.AuthService.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// JWTMiddleware validates the JWT token and attaches claims to the context.
// Requirement 36.3, 36.4: Extract and verify JWT from Authorization header
func (h *Handlers) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := h.AuthService.ValidateToken(parts[1])
		if err != nil {
			if authErr, ok := err.(*model.AuthError); ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": authErr.Message, "code": authErr.Code})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("claims", claims)
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
