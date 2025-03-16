package app

import (
	"shorty/app/routes"
	"shorty/app/routes/ui"
	"shorty/config"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func router(app *fiber.App) {
	// For ping-pong
	app.Get("/ping", func(ctx fiber.Ctx) error { return ctx.SendString("pong") })

	// Init auth & auth store
	ui.InitStore()
	ui.InitOAuth()

	app.Use("/web", static.New("web", static.Config{Compress: true}))
	app.Get("/web/auth/gitlab", ui.OauthLogin)
	app.Get("/web/auth/gitlab/callback", ui.Callback)
	app.Get("/web/*", func(ctx fiber.Ctx) error {
		return ctx.SendFile("web/index.html")
	})

	// UI
	app.Use("/*", static.New("ui", static.Config{
		Compress: true,
		Next: func(ctx fiber.Ctx) bool {
			return strings.HasPrefix(ctx.Path(), "/web")
		},
	}))

	app.Get("/auth/gitlab", ui.OauthLogin)
	app.Get("/auth/gitlab/callback", ui.Callback)
	app.Get("/auth/check", ui.CheckSession)
	app.Get("/login", func(ctx fiber.Ctx) error { return ctx.Render("login", nil) })
	app.Get("/logout", ui.Logout)
	app.Post("/shorty", ui.Create)
	app.Post("/check-filename", ui.CheckFilename)
	app.Get("/events", ui.SSE) // SSE
	app.Patch("/:oldName/:newName", ui.Change)
	app.Delete("/:shorty", ui.Delete)

	if config.Use.S3.Enable {
		app.Post("/upload", ui.Upload)
	}

	// wasm
	// app.Get("/web/*", static.New("web", static.Config{Compress: true}))

	// Get real url
	app.Get("/:shorty", routes.Get)

	// API group
	v1 := app.Group("/v1", verifyKey())
	v1.Post("/shorty", routes.Shorten)             // Create short url
	v1.Delete("/:shorty", routes.Delete)           // Delete url
	v1.Patch("/:oldName/:newName?", routes.Change) // Rename url
	v1.Get("/list", routes.List)                   // List all urls
}
