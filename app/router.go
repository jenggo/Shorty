package app

import (
	"shorty/app/routes"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func router(app *fiber.App) {
	// For ping-pong
	app.Get("/ping", func(ctx *fiber.Ctx) error { return ctx.SendString("pong") })

	// UI
	routes.InitStore()
	routes.InitOAuth()
	app.Static("/", "ui", fiber.Static{Compress: true})
	app.Get("/auth/gitlab", routes.UIOauthLogin)
	app.Get("/auth/gitlab/callback", routes.UICallback)
	app.Get("/auth/check", routes.CheckSession)
	app.Get("/login", func(ctx *fiber.Ctx) error { return ctx.Redirect("/") })
	app.Get("/logout", routes.UILogout)
	app.Get("/ws", routes.Upgrade, websocket.New(routes.Websocket))
	app.Post("/shorty", routes.UICreate)
	app.Patch("/:oldName/:newName", routes.UIChange)
	app.Delete("/:shorty", routes.UIDelete)

	// Get real url
	app.Get("/:shorty", routes.Get)

	// API group
	v1 := app.Group("/v1", verifyKey())
	v1.Post("/shorty", routes.Shorten)             // Create short url
	v1.Delete("/:shorty", routes.Delete)           // Delete url
	v1.Patch("/:oldName/:newName?", routes.Change) // Rename url
	v1.Get("/list", routes.List)                   // List all urls
}
