package main

import (
	"net/http"
	"testing"
)

func TestGetUserHandler(t *testing.T) {
	withRedis := config{
		redisCfg: redisConfig{
			enabled: true,
		},
	}
	app := NewTestApplication(t, withRedis)
	mux := app.mount()
	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatalf("error: %s\n", err.Error())
	}

	t.Run("should not allow unathenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/v1/users/190", nil)
		if err != nil {
			t.Fatalf("error: %s\n", err.Error())
		}

		rr := executeRequest(req, mux)

		// if rr.Code != http.StatusUnauthorized {
		// 	t.Errorf("expected the response code to be %d and got %d", http.StatusUnauthorized, rr.Code)
		// }
		checkResponse(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allowed authencticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/v1/users/190", nil)
		if err != nil {
			t.Fatalf("error: %s\n", err.Error())
		}

		req.Header.Set("Autorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponse(t, http.StatusOK, rr.Code)
	})
}
