package routes

import (
	"fmt"
	"shorty/pkg"
	"shorty/types"

	"github.com/gofiber/fiber/v2"
)

func Change(ctx *fiber.Ctx) error {
	oldName := ctx.Params("oldName")
	urlValue, err := pkg.Redis.Get(ctx.Context(), oldName)
	if err != nil {
		return err
	}

	newName := ctx.Params("newName")

	var body types.Shorten
	if newName == "" {
		if err := ctx.BodyParser(&body); err != nil {
			return err
		}

		if body.Shorty == "" {
			return ctx.JSON(types.Response{
				Error:   true,
				Message: "new shorty cannot be empty",
			})
		}

		newName = body.Shorty
	}

	if newName == oldName {
		return ctx.JSON(types.Response{
			Error:   true,
			Message: "both shorty cannot be the same",
		})
	}

	if err := pkg.Redis.Set(ctx.Context(), newName, urlValue, body.Expired); err != nil {
		return err
	}

	if err := pkg.Redis.Del(ctx.Context(), oldName); err != nil {
		return err
	}

	return ctx.JSON(types.Response{
		Error:   false,
		Message: fmt.Sprintf("%s changed to %s", oldName, newName),
	})
}
