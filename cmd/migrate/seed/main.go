package main

import (
	"log"
	"time"

	"github.com/AlieNoori/social/internal/db"
	"github.com/AlieNoori/social/internal/env"
	"github.com/AlieNoori/social/internal/store"
)

func main() {
	env.LoadEnv(".env")
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable")
	conn, err := db.New(addr, 3, 3, 15*time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	storage := store.NewStorage(conn)

	db.Seed(storage, conn)
}
