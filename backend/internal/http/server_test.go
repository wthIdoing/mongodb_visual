package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mongoDB_visual/backend/internal/config"
	"mongoDB_visual/backend/internal/mongodb"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowsLocalDevelopmentOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewServer(&config.Config{
		FrontendOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
	}, mongodb.NewPool(mongodb.ConnectionConfig{}))

	testCases := []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}

	for _, origin := range testCases {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/databases/test/collections/test/documents/507f1f77bcf86cd799439011", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", http.MethodDelete)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("origin %s returned status %d, want %d", origin, recorder.Code, http.StatusNoContent)
		}
		if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("origin %s returned Access-Control-Allow-Origin %q", origin, got)
		}
	}
}

func TestCORSRejectsUnrelatedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewServer(&config.Config{
		FrontendOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
	}, mongodb.NewPool(mongodb.ConnectionConfig{}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/databases/test/collections/test/documents/507f1f77bcf86cd799439011", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected Access-Control-Allow-Origin header %q", got)
	}
}

func TestConnectionConfigFromRequestUsesHeadersAndFallbacks(t *testing.T) {
	server := &Server{
		pool: mongodb.NewPool(mongodb.ConnectionConfig{
			Host:       "fallback-host",
			Port:       "27018",
			Database:   "fallback-db",
			AuthSource: "fallback-auth",
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/databases", nil)
	req.Header.Set(headerMongoHost, "mongodb.default.svc.cluster.local")
	req.Header.Set(headerMongoPort, "27017")
	req.Header.Set(headerMongoUsername, "tester")
	req.Header.Set(headerMongoPassword, "secret")

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	conn := server.connectionConfigFromRequest(ctx)

	if conn.Host != "mongodb.default.svc.cluster.local" {
		t.Fatalf("host = %q", conn.Host)
	}
	if conn.Port != "27017" {
		t.Fatalf("port = %q", conn.Port)
	}
	if conn.Database != "fallback-db" {
		t.Fatalf("database = %q", conn.Database)
	}
	if conn.Username != "tester" {
		t.Fatalf("username = %q", conn.Username)
	}
	if conn.Password != "secret" {
		t.Fatalf("password = %q", conn.Password)
	}
	if conn.AuthSource != "fallback-auth" {
		t.Fatalf("authSource = %q", conn.AuthSource)
	}
}

func TestClassifyMongoError(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		wantStatus int
		wantSubstr string
	}{
		{
			name:       "authentication",
			err:        errors.New("Authentication failed."),
			wantStatus: http.StatusUnauthorized,
			wantSubstr: "authentication failed",
		},
		{
			name:       "permission",
			err:        errors.New("not authorized on admin to execute command"),
			wantStatus: http.StatusForbidden,
			wantSubstr: "does not have permission",
		},
		{
			name:       "connection",
			err:        errors.New("server selection error: dial tcp 127.0.0.1:27017: connect: connection refused"),
			wantStatus: http.StatusBadGateway,
			wantSubstr: "Unable to connect",
		},
		{
			name:       "generic",
			err:        errors.New("invalid port"),
			wantStatus: http.StatusBadRequest,
			wantSubstr: "invalid port",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, message := classifyMongoError(tc.err)
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
			if message == "" || !strings.Contains(message, tc.wantSubstr) {
				t.Fatalf("message = %q, want substring %q", message, tc.wantSubstr)
			}
		})
	}
}
