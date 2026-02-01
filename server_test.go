package mmdb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	s := &Server{
		AuthToken: "secret",
	}

	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		unauthorized, _ := r.Context().Value("unauthorized").(bool)
		if unauthorized {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized_in_context"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}
	}))

	tests := []struct {
		name           string
		header         string
		query          string
		accept         string
		format         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid Header",
			header:         "Bearer secret",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "Valid Query",
			query:          "auth=secret",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "Invalid Token - JSON",
			header:         "Bearer wrong",
			accept:         "application/json",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "Invalid Token - Text",
			header:         "Bearer wrong",
			accept:         "text/plain",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "Invalid Token - HTML (should pass through with context)",
			header:         "Bearer wrong",
			accept:         "text/html",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "unauthorized_in_context",
		},
		{
			name:           "No Token - HTML (should pass through with context)",
			accept:         "text/html",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "unauthorized_in_context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/"
			if tt.query != "" {
				url += "?" + tt.query
			}
			if tt.format != "" {
				if tt.query == "" {
					url += "?"
				} else {
					url += "&"
				}
				url += "format=" + tt.format
			}
			req := httptest.NewRequest("GET", url, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
			if tt.expectedStatus == http.StatusUnauthorized && strings.Contains(rr.Body.String(), s.AuthToken) {
				t.Errorf("body must not contain auth token if unauthorized")
			}
		})
	}
}
