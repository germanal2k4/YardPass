package middleware

import (
	"context"
	"strings"

	"yardpass/internal/auth"
	"yardpass/internal/errors"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errors.Unauthorized(c, "MISSING_TOKEN", "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errors.Unauthorized(c, "INVALID_TOKEN_FORMAT", "Token must be in format: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(c.Request.Context(), parts[1])
		if err != nil {
			errors.Unauthorized(c, "INVALID_TOKEN", "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		if claims.BuildingID != nil {
			c.Set("building_id", *claims.BuildingID)
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "building_id", *claims.BuildingID))
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "user_id", claims.UserID))
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "role", claims.Role))

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			errors.Unauthorized(c, "MISSING_ROLE", "User role not found")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			errors.Unauthorized(c, "INVALID_ROLE", "Invalid role type")
			c.Abort()
			return
		}

		for _, r := range roles {
			if roleStr == r {
				c.Next()
				return
			}
		}

		errors.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Insufficient permissions")
		c.Abort()
	}
}

func ServiceAuthMiddleware(serviceToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Service-Token")
		if token == "" {
			errors.Unauthorized(c, "MISSING_SERVICE_TOKEN", "Service token is required")
			c.Abort()
			return
		}

		if token != serviceToken {
			errors.Unauthorized(c, "INVALID_SERVICE_TOKEN", "Invalid service token")
			c.Abort()
			return
		}

		c.Next()
	}
}
