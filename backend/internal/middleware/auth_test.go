package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type denyAllPermissionChecker struct{}

func (d *denyAllPermissionChecker) HasPermission(role domain.Role, featureKey string) (bool, error) {
	if role == "admin" || role == "superadmin" {
		return false, fmt.Errorf("CRITICAL BUG: HasPermission called for admin role when it should have short-circuited!")
	}
	return false, nil
}

func TestRequirePermission_AdminAlwaysAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &denyAllPermissionChecker{}
	mw := middleware.RequirePermission(checker, "devices.edit")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserRole, "admin")
		c.Next()
	})
	router.GET("/test-admin", mw, func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-admin", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected admin role to pass RequirePermission with 200 OK, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequirePermission_AnggotaDeniedWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &denyAllPermissionChecker{}
	mw := middleware.RequirePermission(checker, "devices.edit")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserRole, "anggota")
		c.Next()
	})
	router.GET("/test-anggota", mw, func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-anggota", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected anggota role to be denied with 403 Forbidden, got %d: %s", w.Code, w.Body.String())
	}
}
