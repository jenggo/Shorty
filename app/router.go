package app

import (
	"shorty/app/routes"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func router(app *fiber.App) {
	// For ping-pong
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// UI
	app.Static("/css", "ui/css", fiber.Static{Compress: true}) // For static css/js
	app.Get("/", routes.HTMLMain)
	app.Post("/login", routes.HTMLLogin)
	app.Get("/logout", routes.HTMLLogout)
	app.Get("/ws", routes.Upgrade, websocket.New(routes.Websocket))

	// GetURL
	app.Get("/:shorty", routes.Get)

	// API group
	v1 := app.Group("/v1", verifyKey())
	v1.Post("/shorty", routes.Shorten)             // Create short url
	v1.Delete("/:shorty", routes.Delete)           // Delete url
	v1.Patch("/:oldName/:newName?", routes.Change) // Rename url
	v1.Get("/list", routes.List)                   // List all urls
}
