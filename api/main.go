package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}
func main() {
	//load the environment variable from the .env file
	err := godotenv.Load()

	if err != nil {
		fmt.Println(err)
	}

	// initialize the app of the fiber router
	app := fiber.New()

	app.Use(logger.New())

	//setting up the routes
	setupRoutes(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))

}
