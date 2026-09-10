package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		checkHeaders   func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "allowed origin sets CORS headers",
			origin:         "http://localhost:3000",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
					t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
				}
				if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
					t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
				}
				if got := w.Header().Get("Vary"); got != "Origin" {
					t.Errorf("Vary = %q, want %q", got, "Origin")
				}
			},
		},
		{
			name:           "disallowed origin does not set CORS headers",
			origin:         "http://evil.example.com",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
					t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
				}
				if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
					t.Errorf("Access-Control-Allow-Credentials = %q, want empty", got)
				}
			},
		},
		{
			name:           "OPTIONS preflight returns 204 with CORS headers",
			origin:         "http://localhost:3000",
			method:         http.MethodOptions,
			expectedStatus: http.StatusNoContent,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
					t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
				}
				if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
					t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
				}
				if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
					t.Error("Access-Control-Allow-Methods should not be empty")
				}
				if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" {
					t.Error("Access-Control-Allow-Headers should not be empty")
				}
				if got := w.Header().Get("Access-Control-Max-Age"); got != "86400" {
					t.Errorf("Access-Control-Max-Age = %q, want %q", got, "86400")
				}
			},
		},
		{
			name:           "OPTIONS preflight from disallowed origin returns 204 without CORS headers",
			origin:         "http://evil.example.com",
			method:         http.MethodOptions,
			expectedStatus: http.StatusNoContent,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
					t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
				}
			},
		},
		{
			name:           "no origin header does not set CORS headers",
			origin:         "",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
					t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORSMiddleware([]string{"http://localhost:3000"}))
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
			router.OPTIONS("/test", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.checkHeaders != nil {
				tt.checkHeaders(t, w)
			}
		})
	}
}

func TestCORSMiddlewareAllowsAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORSMiddleware([]string{"http://localhost:3000"}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Authorization", "Bearer test-token")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
	}
}
