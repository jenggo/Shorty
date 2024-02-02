package app

import (
	"errors"
	"os"
	"shorty/config"
	"shorty/types"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/template/html/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func RunServer() (app *fiber.App, err error) {
	appCfg := fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
		ErrorHandler:          errHandler,
		ReadTimeout:           10 * time.Second,
		ProxyHeader:           "Cf-Connecting-Ip",
		Views:                 html.New("ui", ".tpl"),
	}

	if !config.Use.App.Cloudflare {
		appCfg.ProxyHeader = "X-Real-Ip"
	}

	app = fiber.New(appCfg)

	if config.Use.App.PPROF != "" {
		log.Log().Msgf("» pprof enabled: %s", config.Use.App.PPROF)
		app.Use(pprof.New(pprof.Config{Prefix: config.Use.App.PPROF}))
	}

	app.Use(cors.New())
	app.Use(favicon.New())
	app.Use(logger.New(loggerConfig()))
	app.Use(helmet.New())
	app.Use(earlydata.New())
	app.Use(etag.New())
	router(app)

	go func() {
		log.Log().Msgf("» %s %s listen: %s", config.AppName, config.AppVersion, config.Use.App.Listen)

		if err := app.Listen(config.Use.App.Listen); err != nil {
			log.Fatal().Caller().Err(err).Send()
		}
	}()

	return
}

func errHandler(c *fiber.Ctx, err error) error {
	statusCode := c.Response().StatusCode()
	if statusCode == fiber.StatusNotFound || statusCode == fiber.StatusOK {
		return nil
	}

	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	ua := c.Get(fiber.HeaderUserAgent)
	ip := getIP(c)
	method := c.Method()
	path := c.Path()

	if ua != "" && ip != "" {
		log.Error().Str("UserAgent", ua).Str("IP", ip).Str("Method", method).Str("Path", path).Err(err).Send()
	}

	return c.Status(code).JSON(types.Response{
		Error:   true,
		Message: err.Error(),
	})
}

func loggerConfig() (cfg logger.Config) {
	format := "» ${time} ${method} ${blue}${path}${reset} [${status}] [${red}${ip}${reset}] [${reqHeader:Accept-Encoding}] [${latency}] — ${magenta}${ua}${reset}\n"
	level := config.Use.App.LogLevel

	if level < 3 {
		format += "» ${blue}${reqHeader:Authorization}${reset}\n» Error: ${red}${error}${reset}\n» Header: ${cyan}${reqHeaders}${reset}\n"

		if level < 2 {
			format += "» Body: ${body}\n\n"
		}
	}

	return logger.Config{
		Next: func(c *fiber.Ctx) bool {
			statusCode := c.Response().StatusCode()
			return statusCode == fiber.StatusNotFound || statusCode == fiber.StatusOK
		},
		Format:     format,
		TimeFormat: "2006-01-02T15:04:05",
		Output:     os.Stderr, // If using os.Stdout, log does not colorize
	}
}

func getIP(c *fiber.Ctx) (ip string) {
	ipn := strings.TrimSpace(c.IP())
	ips := c.IPs()
	cf := strings.TrimSpace(c.Get("Cf-Connecting-Ip"))
	xr := strings.TrimSpace(c.Get("X-Real-Ip"))

	logs := log.Sample(zerolog.Sometimes)
	logs.Debug().Msgf("IP: %s, IPs: %v, Cf-Connecting-Ip: %s, X-Real-Ip: %s", ip, ips, cf, xr)

	switch {
	case ipn != "":
		ip = ipn
	case cf != "":
		ip = cf
	case xr != "":
		ip = xr
	case len(ips) > 0:
		ip = strings.Join(ips, ",")
	}

	if ip == "" {
		logs.Error().Msg("No IP detected")
	}

	return
}
