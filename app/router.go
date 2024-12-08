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
	app.Static("/", "ui", fiber.Static{Compress: true})
	app.Post("/login", routes.UILogin)
	app.Get("/logout", routes.UILogout)
	app.Delete("/:shorty", routes.UIDelete)
	app.Patch("/:oldName/:newName", routes.UIChange)
	app.Post("/shorty", routes.UICreate)
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
