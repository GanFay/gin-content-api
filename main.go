package main

import (
	"blog/db"
	"blog/handlers"
	"blog/repository"
	"blog/router"
	"log"
	"os"

	_ "blog/docs"

	"github.com/joho/godotenv"
)

// @title           Gin Content API
// @version         1.0
// @description     A robust, production-ready RESTful API designed for content management.
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.
func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is empty")
	}

	pool := db.MustConnect(dbURL)
	defer pool.Close()

	postRep := repository.NewPostRepository(pool)
	userRep := repository.NewUserRepository(pool)

	h := handlers.NewHandler(postRep, userRep)
	r := router.SetupRouter(h)
	err := r.Run(":8080")
	if err != nil {
		return
	}
}
