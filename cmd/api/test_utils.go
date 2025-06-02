package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlieNoori/social/internal/auth"
	"github.com/AlieNoori/social/internal/ratelimiter"
	"github.com/AlieNoori/social/internal/store"
	"github.com/AlieNoori/social/internal/store/cache"
	"go.uber.org/zap"
)

func NewTestApplication(t *testing.T, cfg config) *application {
	t.Helper()

	// logger := zap.Must(zap.NewProduction()).Sugar()
	logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()
	testAuth := &auth.TestAuthenticator{}

	// Rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	return &application{
		logger:        logger,
		store:         mockStore,
		config:        cfg,
		cacheStore:    mockCacheStore,
		rateLimiter:   rateLimiter,
		authenticator: testAuth,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponse(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected the response code to be %d and got %d", expected, actual)
	}
}
