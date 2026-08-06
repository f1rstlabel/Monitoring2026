package middleware

import (
	"net/http"
	"strings"

	"govmonitor-it/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("govmonitor-secret-key-2.4.1")

// contextKey constants for values injected into Gin context.
const (
	ContextUserID   = "userID"
	ContextUserRole = "userRole"
	ContextUserName = "userName"
)

// JWTMiddleware validates the Bearer token and injects user claims into context.
// Routes without a token will get a 401.
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Malformed Authorization header — expected 'Bearer <token>'",
			})
			return
		}

		tokenStr := parts[1]

		// Accept demo tokens for local dev without signature verification.
		// Format: demo-jwt-<role>-<version>
		if strings.HasPrefix(tokenStr, "demo-jwt-") {
			segments := strings.Split(tokenStr, "-")
			// demo-jwt-superadmin-2.4.1  →  segments[2] = superadmin
			if len(segments) >= 3 {
				role := segments[2]
				c.Set(ContextUserID, "demo-user")
				c.Set(ContextUserRole, role)
				c.Set(ContextUserName, "Demo "+strings.Title(role))
				c.Next()
				return
			}
		}

		// Standard JWT validation.
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		c.Set(ContextUserID, claims["sub"])
		c.Set(ContextUserRole, claims["role"])
		c.Set(ContextUserName, claims["name"])
		c.Next()
	}
}

// RequireRole returns a middleware that allows only the specified roles.
// Role values must match the domain.Role constants: "superadmin", "pimpinan", "anggota".
// Returns 403 Forbidden for valid but insufficient role — NOT 401, to distinguish auth vs authz.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid role format in token",
			})
			return
		}

		// superadmin implicitly passes all role checks across all route groups
		if roleStr == "superadmin" {
			c.Next()
			return
		}

		if _, allowed := roleSet[roleStr]; !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient permissions",
				"required_role": allowedRoles,
				"your_role":     roleStr,
			})
			return
		}

		c.Next()
	}
}

type PermissionChecker interface {
	HasPermission(role string, featureKey string) (bool, error)
}

type DomainRolePermissionChecker interface {
	HasPermission(role domain.Role, featureKey string) (bool, error)
}

// RequirePermission returns a middleware checking the dynamic role_permissions table.
func RequirePermission(checker DomainRolePermissionChecker, featureKey string, fallbackRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}
		roleStr, _ := roleVal.(string)
		if roleStr == "superadmin" {
			c.Next()
			return
		}

		if checker != nil {
			allowed, err := checker.HasPermission(domain.Role(roleStr), featureKey)
			if err == nil {
				if !allowed {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"error":       "Permission denied for feature: " + featureKey,
						"feature_key": featureKey,
						"your_role":   roleStr,
					})
					return
				}
				c.Next()
				return
			}
		}

		if len(fallbackRoles) > 0 {
			for _, r := range fallbackRoles {
				if r == roleStr {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":     "Permission denied for feature: " + featureKey,
			"your_role": roleStr,
		})
	}
}

// GetRole extracts the role string from Gin context (set by JWTMiddleware).
func GetRole(c *gin.Context) string {
	role, _ := c.Get(ContextUserRole)
	s, _ := role.(string)
	return s
}

// GetUserID extracts the user ID from Gin context.
func GetUserID(c *gin.Context) string {
	uid, _ := c.Get(ContextUserID)
	s, _ := uid.(string)
	return s
}
