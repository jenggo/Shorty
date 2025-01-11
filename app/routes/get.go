package routes

import (
	"shorty/pkg"

	"github.com/gofiber/fiber/v2"
)

func Get(ctx *fiber.Ctx) error {
	shorturl := ctx.Params("shorty")

	realurl, err := pkg.Redis.Get(ctx.Context(), shorturl)
	if err != nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	return ctx.Redirect(realurl, fiber.StatusPermanentRedirect)
}
