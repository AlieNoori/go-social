package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlieNoori/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Autorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}
		token := parts[1]

		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			log.Println("validate token error", err)
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)
		userId, err := strconv.Atoi(fmt.Sprintf("%.f", claims["sub"]))
		if err != nil {
			log.Println("userid error", err)
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.getUser(ctx, userId)
		if err != nil {
			log.Println("get user", err)
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtxKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
				return
			}

			// parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}

			// decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
				return
			}

			// check the credentials
			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		post := getPostFromCtx(r)

		if post.UserId == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, userId int) (*store.User, error) {
	if !app.config.redisCfg.enabled {
		user, err := app.store.Users.GetById(ctx, userId)
		if err != nil {
			return nil, err
		}

		return user, nil
	}
	user, err := app.cacheStore.Users.Get(ctx, userId)
	if user == nil || err != nil {
		user, err = app.store.Users.GetById(ctx, userId)
		if err != nil {
			return nil, err
		}

		if err := app.cacheStore.Users.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled && app.config.redisCfg.enabled {
			ctx := r.Context()
			ip := r.RemoteAddr
			window := app.config.rateLimiter.TimeFrame

			count, err := app.cacheStore.RateLimit.Get(ctx, ip)
			if err != nil {
				app.internalServerError(w, r, err)
				return
			}

			if count == 0 {
				if err := app.cacheStore.RateLimit.Set(ctx, ip, window); err != nil {
					app.internalServerError(w, r, err)
					return
				}
				next.ServeHTTP(w, r)
			}

			if count != 0 && count < app.config.rateLimiter.RequestPerTimeFrame {
				if err := app.cacheStore.RateLimit.Incrementor(ctx, ip); err != nil {
					app.internalServerError(w, r, err)
					return
				}
				next.ServeHTTP(w, r)
			}

			app.rateLimitExceededResponse(w, r, window.String())
			return
		}

		if app.config.rateLimiter.Enabled && !app.config.redisCfg.enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
