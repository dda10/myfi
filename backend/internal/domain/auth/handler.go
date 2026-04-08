package auth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handlers holds auth domain dependencies for HTTP handler methods.
type Handlers struct {
	AuthService *AuthService
}

// HandleRegister serves POST /api/auth/register — create a new user account.
func (h *Handlers) HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.AuthService.Register(req)
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			if authErr.Code == ErrCodeUserExists {
				c.JSON(http.StatusConflict, gin.H{"error": authErr.Message, "code": authErr.Code})
				return
			}
		}
		slog.Error("registration failed", "err", err, "username", req.Username)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

// HandleLogin serves POST /api/auth/login — authenticate and issue JWT.
func (h *Handlers) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.AuthService.Login(req)
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			switch authErr.Code {
			case ErrCodeInvalidCredentials:
				c.JSON(http.StatusUnauthorized, gin.H{"error": authErr.Message, "code": authErr.Code})
			case ErrCodeAccountLocked:
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

// HandleLogout serves POST /api/auth/logout — invalidate session.
func (h *Handlers) HandleLogout(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		return
	}
	jwtClaims := claims.(*JWTClaims)
	_ = h.AuthService.Logout(jwtClaims.SessionID)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// HandleChangePassword serves POST /api/auth/change-password.
func (h *Handlers) HandleChangePassword(c *gin.Context) {
	userID := getUserID(c)
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.AuthService.ChangePassword(userID, req); err != nil {
		if authErr, ok := err.(*AuthError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": authErr.Message, "code": authErr.Code})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password change failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// HandleMe serves GET /api/auth/me — get current user info.
func (h *Handlers) HandleMe(c *gin.Context) {
	userID := getUserID(c)
	user, err := h.AuthService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// JWTMiddleware validates the JWT token and attaches claims to the context.
func (h *Handlers) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := h.AuthService.ValidateToken(parts[1])
		if err != nil {
			if authErr, ok := err.(*AuthError); ok {
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

// --- Helpers ---

func getUserID(c *gin.Context) string {
	if id, exists := c.Get("userID"); exists {
		if v, ok := id.(string); ok {
			return v
		}
	}
	return ""
}
