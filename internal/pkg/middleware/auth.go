package middleware

import (
	"strings"

	"github.com/akza/akza-api/internal/pkg/apperror"
	jwtpkg "github.com/akza/akza-api/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const AdminIDKey = "admin_id"

// AuthRequired validates the Bearer JWT and injects admin_id into context.
func AuthRequired(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			respond401(c)
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respond401(c)
			return
		}
		claims, err := jwtManager.ParseClaims(parts[1])
		if err != nil {
			respond401(c)
			return
		}
		c.Set(AdminIDKey, claims.AdminID)
		c.Set("admin_email", claims.Email)
		c.Next()
	}
}

func respond401(c *gin.Context) {
	c.AbortWithStatusJSON(401, gin.H{"error": apperror.ErrUnauthorized})
}
