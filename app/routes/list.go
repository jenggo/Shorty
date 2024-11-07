package routes

import (
	"shorty/pkg"
	"shorty/types"

	"github.com/gofiber/fiber/v2"
)

func List(ctx *fiber.Ctx) error {
	list, err := pkg.Redis.GetAll(ctx.Context())
	if err != nil {
		return err
	}

	// re-check
	if len(list) == 0 {
		return ctx.Status(404).JSON(types.Response{
			Error:   true,
			Message: "still empty",
		})
	}

	return ctx.JSON(list)
}
