package app

import (
	"strings"

	"shorty/app/routes"
	"shorty/app/routes/ui"
	"shorty/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func router(app *fiber.App) {
	// For ping-pong
	app.Get("/ping", func(ctx fiber.Ctx) error { return ctx.SendString("pong") })

	// Init auth & auth store
	ui.InitStore()
	ui.InitOAuth()

	// Per-IP limits, built once so each limiter shares one storage-backed
	// counter set. Each scope gets its own key prefix.
	redirectLimit := rateLimiter(ui.RedisStorage, "redirect", 100)
	createLimit := rateLimiter(ui.RedisStorage, "create", 10)
	writeLimit := rateLimiter(ui.RedisStorage, "write", 20)
	listLimit := rateLimiter(ui.RedisStorage, "list", 20)
	authLimit := rateLimiter(ui.RedisStorage, "auth", 5)
	uploadLimit := rateLimiter(ui.RedisStorage, "upload", 20)
	checkLimit := rateLimiter(ui.RedisStorage, "check", 60)

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

	app.Get("/auth/gitlab", authLimit, ui.OauthLogin)
	app.Get("/auth/gitlab/callback", authLimit, ui.Callback)
	app.Get("/auth/config", ui.GetAuthConfig)
	app.Post("/auth/login", authLimit, ui.LoginUserPass)
	app.Get("/auth/check", ui.CheckSession)
	app.Get("/login", func(ctx fiber.Ctx) error { return ctx.Render("login", nil) })
	app.Get("/logout", ui.Logout)
	app.Post("/shorty", createLimit, ui.Create)
	app.Post("/check-filename", checkLimit, ui.CheckFilename)
	app.Get("/events", ui.SSE) // SSE
	app.Patch("/:oldName/:newName", writeLimit, ui.Change)
	app.Delete("/:shorty", writeLimit, ui.Delete)

	if config.Use.S3.Enable {
		app.Post("/upload", uploadLimit, ui.Upload)
	}

	// wasm
	// app.Get("/web/*", static.New("web", static.Config{Compress: true}))

	// Get real url
	app.Get("/:shorty", redirectLimit, routes.Get)

	// API group
	v1 := app.Group("/v1", verifyKey())
	v1.Post("/shorty", createLimit, routes.Shorten)            // Create short url
	v1.Delete("/:shorty", writeLimit, routes.Delete)           // Delete url
	v1.Patch("/:oldName/:newName?", writeLimit, routes.Change) // Rename url
	v1.Get("/list", listLimit, routes.List)                    // List all urls
}
