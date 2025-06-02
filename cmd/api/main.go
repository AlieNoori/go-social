package main

import (
	"expvar"
	"log"
	"runtime"
	"time"

	"github.com/AlieNoori/social/internal/auth"
	"github.com/AlieNoori/social/internal/db"
	"github.com/AlieNoori/social/internal/env"
	"github.com/AlieNoori/social/internal/mailer"
	"github.com/AlieNoori/social/internal/ratelimiter"
	"github.com/AlieNoori/social/internal/store"
	"github.com/AlieNoori/social/internal/store/cache"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const version string = "0.0.1"

//	@title			Gopher Social API
//	@description	This is a API for Gopher Social
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/v1
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Autorization
//	@description

func main() {
	err := env.LoadEnv(".env")
	if err != nil {
		log.Fatalf("can not load environment variables: %s", err)
	}

	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		},
		redisCfg: redisConfig{
			addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PASSWORD", ""),
			db:       env.GetInt("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", true),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       env.GetDuration("EXPIRE_INVITE_EMAIL", time.Hour*24*3),
			fromEmail: env.GetString("FROME_MAIL", ""),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		frontendURL: env.GetString("FRONEND_URL", "http://localhost:4000"),
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3,
				iss:    "gophersocial",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_COUNT", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Info("Database connection pool established")

	// cache
	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb, err = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.password, cfg.redisCfg.db)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("Redis cache connection established")
		defer rdb.Close()
	}

	store := store.NewStorage(db)
	cacheStore := cache.NewRedisStorage(rdb)

	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	// mailer := mailer.NewSendGrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)
	mailer := mailer.NewSendGrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	app := &application{
		config:        cfg,
		store:         store,
		cacheStore:    cacheStore,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
	}

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any { return db.Stats() }))
	expvar.Publish("go-routins", expvar.Func(func() any { return runtime.NumGoroutine() }))

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
