package routes

import (
	"shorty/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Get(ctx *fiber.Ctx) error {
	shorturl := ctx.Params("shorty")

	realurl, err := pkg.Redis.Get(shorturl)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	if realurl == "" {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	return ctx.Redirect(realurl, fiber.StatusPermanentRedirect)
}
