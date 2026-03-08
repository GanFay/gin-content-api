package main

import (
	"blog/handlers"
	"blog/router"
	"context"
	"log"
	"os"

	_ "blog/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// @title Blogging Platform API
// @version 1.0
// @description REST API for a blogging platform built with Go, Gin, PostgreSQL, JWT auth, refresh tokens and ownership protection.
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is empty")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connected DB: ", err)
	}
	defer pool.Close()

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS posts (
					id         SERIAL PRIMARY KEY,
					author_id  TEXT	NOT NULL,
					title      TEXT UNIQUE NOT NULL,
					content    TEXT        NOT NULL,
					category   TEXT,
					tags TEXT[],
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		    		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
		CREATE TABLE IF NOT EXISTS users (
		    id         SERIAL PRIMARY KEY,
		    username   TEXT UNIQUE NOT NULL,
		    email      TEXT UNIQUE NOT NULL,
		    password_hash TEXT NOT NULL,
		    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);			
`)

	if err != nil {
		log.Fatal("Create table error:", err)
	}
	log.Println("Table is ready")

	h := handlers.NewHandler(pool)
	r := router.SetupRouter(h)
	err = r.Run(":8080")
	if err != nil {
		return
	}
}
