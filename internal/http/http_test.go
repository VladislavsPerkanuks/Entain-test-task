package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET root endpoint",
			method:         "GET",
			url:            "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "welcome",
		},
		{
			name:           "GET non-existent endpoint",
			method:         "GET",
			url:            "/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			require.NoError(t, err, "Could not create request")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "Status code")
			assert.Equal(t, tt.expectedBody, rr.Body.String(), "Response body")
		})
	}
}

func TestHealthCheck(t *testing.T) {
	router := NewRouter()

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err, "Could not create request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Health check status")
	assert.Equal(t, "welcome", rr.Body.String(), "Health check response")
}
