package routes

import (
	"shorty/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Get(ctx *fiber.Ctx) error {
	shorturl := ctx.Params("shorty")

	realurl, err := pkg.Redis.Get(ctx.Context(), shorturl)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	return ctx.Redirect(realurl, fiber.StatusPermanentRedirect)
}
