package app

import (
	"errors"
	"shorty/config"
	"shorty/types"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/earlydata"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/rs/zerolog/log"
)

func RunServer() (app *fiber.App, err error) {
	appCfg := fiber.Config{
		AppName:       config.AppName,
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
		ErrorHandler:  errHandler,
		ProxyHeader:   "Cf-Connecting-Ip",
		Views:         html.New("ui", ".html"),
		CaseSensitive: true,
		ReadTimeout:   10 * time.Second,
	}

	if config.Use.S3.Enable {
		appCfg.StreamRequestBody = true
		appCfg.BodyLimit = 100 * 1024 * 1024
		appCfg.ReadTimeout = time.Minute
	}

	if !config.Use.App.Cloudflare {
		appCfg.ProxyHeader = "X-Real-Ip"
	}

	app = fiber.New(appCfg)

	if config.Use.App.PPROF != "" {
		log.Log().Msgf("» pprof enabled: %s", config.Use.App.PPROF)
		app.Use(pprof.New(pprof.Config{Prefix: config.Use.App.PPROF}))
	}

	// storage := redis.New(redis.Config{
	// 	Host:     config.Use.Redis.Host,
	// 	Password: config.Use.Redis.Password,
	// 	Database: config.Use.Redis.DB.Auth + 1,
	// })

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.Use.App.BaseURL},
		AllowHeaders:     []string{"Origin, Content-Type, Accept, Authorization, Cache-Control"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	app.Use(favicon.New())
	app.Use(helmet.New())
	app.Use(earlydata.New())
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	// app.Use(limiter.New(limiter.Config{
	// 	Expiration:             5 * time.Minute,
	// 	LimiterMiddleware:      limiter.SlidingWindow{},
	// 	SkipSuccessfulRequests: true,
	// 	Storage:                storage,
	// 	Next: func(c fiber.Ctx) bool {
	// 		switch c.Path() {
	// 		case "/login":
	// 			return true
	// 		case "/web":
	// 			return true
	// 		case "/_app":
	// 			return true
	// 		default:
	// 			return false
	// 		}
	// 	},
	// }))

	router(app)

	go func() {
		log.Log().Msgf("» %s %s listen: %s", config.AppName, config.AppVersion, config.Use.App.Listen)

		if err := app.Listen(config.Use.App.Listen, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
			log.Error().Caller().Err(err).Send()
		}
	}()

	return
}

func errHandler(c fiber.Ctx, err error) error {
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
