package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"sanoc/backend/internal/config"
	"sanoc/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey constants for values injected into Gin context.
const (
	ContextUserID   = "userID"
	ContextUserRole = "userRole"
	ContextUserName = "userName"
)

// SecurityHeaders injects OWASP recommended HTTP security headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.google.com https://www.gstatic.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; img-src 'self' data: https: blob:; connect-src 'self' ws: wss: https:;")
		c.Next()
	}
}

// rateLimiter tracking structure for login endpoint
type loginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

var loginLimiter = &loginRateLimiter{
	attempts: make(map[string][]time.Time),
}

func (l *loginRateLimiter) RecordFailedAttempt(ip string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	var valid []time.Time
	for _, t := range l.attempts[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	valid = append(valid, now)
	l.attempts[ip] = valid
	return len(valid)
}

func RecordFailedLoginAttempt(ip string) int {
	return loginLimiter.RecordFailedAttempt(ip)
}

// RateLimitLogin prevents brute-force attempts on login endpoint.
func RateLimitLogin(maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if clientIP == "" {
			clientIP = c.Request.RemoteAddr
		}

		loginLimiter.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-window)

		// Filter out old timestamps
		var valid []time.Time
		for _, t := range loginLimiter.attempts[clientIP] {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}

		if len(valid) >= maxAttempts {
			loginLimiter.attempts[clientIP] = valid
			loginLimiter.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many login attempts. Please wait a minute before trying again.",
				"retryIn": 60,
			})
			return
		}

		loginLimiter.attempts[clientIP] = append(valid, now)
		loginLimiter.mu.Unlock()
		c.Next()
	}
}

// CSRFProtection verifies X-CSRF-Token for state-changing requests when session cookie is present.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {
			if _, err := c.Cookie("sanoc_session"); err == nil {
				csrfHeader := c.GetHeader("X-CSRF-Token")
				if csrfHeader == "" {
					csrfHeader = c.GetHeader("X-Requested-With")
				}
				if csrfHeader == "" {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"error": "CSRF token missing or invalid",
					})
					return
				}
			}
		}
		c.Next()
	}
}

// JWTMiddleware validates the Bearer token and injects user claims into context.
// Routes without a token will get a 401.
func JWTMiddleware() gin.HandlerFunc {
	cfg := config.LoadConfig()
	jwtSecret := []byte(cfg.JWTSecret)

	return func(c *gin.Context) {
		var tokenStr string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr = parts[1]
			}
		}

		if tokenStr == "" {
			if cookie, err := c.Cookie("sanoc_session"); err == nil && cookie != "" {
				tokenStr = cookie
			}
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required — missing session token or authorization header",
			})
			return
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
				"error": "Invalid or expired session token",
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

		role, _ := claims["role"].(string)
		if role == "superadmin" {
			role = "admin"
		}

		c.Set(ContextUserID, claims["sub"])
		c.Set(ContextUserRole, role)
		c.Set(ContextUserName, claims["name"])
		c.Next()
	}
}

// RequireRole returns a middleware that allows only the specified roles.
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

		// admin / superadmin implicitly passes all role checks across all route groups
		if roleStr == "admin" || roleStr == "superadmin" {
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
// CRITICAL SECURITY: Admin role MUST short-circuit to c.Next() without checking role_permissions!
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

		// Top admin role ALWAYS passes permission checks immediately without querying DB matrix
		if roleStr == "admin" || roleStr == "superadmin" {
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
				if r == roleStr || (r == "admin" && roleStr == "superadmin") {
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
