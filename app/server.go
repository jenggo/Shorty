package app

import (
	"errors"
	"shorty/config"
	"shorty/types"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/keyauth/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/rs/zerolog/log"
)

func RunServer() (app *fiber.App, err error) {
	appCfg := fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
		ErrorHandler:          errHandler,
		ProxyHeader:           "Cf-Connecting-Ip",
		Views:                 html.New("ui", ".html"),
	}

	if config.Use.S3.Enable {
		appCfg.StreamRequestBody = true
		appCfg.BodyLimit = -1
	} else {
		appCfg.ReadTimeout = 10 * time.Second
	}

	if !config.Use.App.Cloudflare {
		appCfg.ProxyHeader = "X-Real-Ip"
	}

	app = fiber.New(appCfg)

	if config.Use.App.PPROF != "" {
		log.Log().Msgf("» pprof enabled: %s", config.Use.App.PPROF)
		app.Use(pprof.New(pprof.Config{Prefix: config.Use.App.PPROF}))
	}

	app.Use(cors.New(cors.Config{
		// AllowOriginsFunc: func(origin string) bool { return true }, // debugging only
		AllowOrigins:     "https://u.nusatek.dev",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Cache-Control",
		AllowCredentials: true,
		MaxAge:           300,
	}))
	app.Use(favicon.New())
	app.Use(helmet.New())
	app.Use(earlydata.New())
	// app.Use(etag.New()) // --> SSE does not work if it enable
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	router(app)

	go func() {
		log.Log().Msgf("» %s %s listen: %s", config.AppName, config.AppVersion, config.Use.App.Listen)

		if err := app.Listen(config.Use.App.Listen); err != nil {
			log.Error().Caller().Err(err).Send()
		}
	}()

	return
}

func errHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	ua := c.Get(fiber.HeaderUserAgent)
	ip := c.IP()
	method := c.Method()
	path := c.Path()

	if ua != "" && ip != "" && code != fiber.StatusNotFound && code != fiber.StatusMethodNotAllowed && err != keyauth.ErrMissingOrMalformedAPIKey {
		log.Error().Str("UserAgent", ua).Str("IP", ip).Str("Method", method).Str("Path", path).Err(err).Send()
	}

	return c.Status(code).JSON(types.Response{
		Error:   true,
		Message: err.Error(),
	})
}
