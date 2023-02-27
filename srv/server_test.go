package srv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUnknownRoute(t *testing.T) {
	hs := Server{
		Logger:          zap.NewNop(),
		AuditFileWriter: &strings.Builder{},
	}
	s := hs.NewServer()
	router := s.Handler

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.TODO(), "GET", "/a/route/that/doesnt/exist", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.Equal(t, `{"message":"invalid request - route not found"}`, w.Body.String())
}

func TestHealthzRoute(t *testing.T) {
	hs := Server{
		Logger:          zap.NewNop(),
		AuditFileWriter: &strings.Builder{},
	}
	s := hs.NewServer()
	router := s.Handler

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.TODO(), "GET", "/healthz", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"status":"UP"}`, w.Body.String())
}

func TestLivenessRoute(t *testing.T) {
	hs := Server{
		Logger:          zap.NewNop(),
		AuditFileWriter: &strings.Builder{},
	}
	s := hs.NewServer()
	router := s.Handler

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.TODO(), "GET", "/healthz/liveness", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"status":"UP"}`, w.Body.String())
}

func TestReadinessRouteUp(t *testing.T) {
	hs := Server{
		Logger:          zap.NewNop(),
		AuditFileWriter: &strings.Builder{},
	}
	s := hs.NewServer()
	router := s.Handler

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.TODO(), "GET", "/healthz/readiness", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"status":"UP"}`, w.Body.String())
}
