package ui

import (
	"shorty/app/routes"
	"shorty/types"

	"github.com/gofiber/fiber/v3"
)

func Change(ctx fiber.Ctx) error {
	if _, err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	return routes.Change(ctx)
}
