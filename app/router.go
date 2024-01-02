package app

import (
	"shorty/app/routes"

	"github.com/gofiber/fiber/v2"
)

func router(app *fiber.App) {
	// For ping-pong
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// GetURL
	app.Get("/:shorty", routes.Get)

	// API group
	v1 := app.Group("/v1", verifyKey())
	v1.Post("/shorty", routes.Shorten)   // Create short url
	v1.Delete("/:shorty", routes.Delete) // Delete url
	v1.Patch("/:shorty", routes.Change)  // Edit url
}
