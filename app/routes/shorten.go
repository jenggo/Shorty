package routes

import (
	"fmt"

	"shorty/pkg"
	"shorty/types"

	"github.com/gofiber/fiber/v2"
)

func Shorten(ctx *fiber.Ctx) error {
	var body types.Shorten
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.Url == "" {
		return fmt.Errorf("url cannot be empty")
	}

	// Check that url
	testUrl := fiber.Head(body.Url)
	statusCode, _, errs := testUrl.Bytes()
	if len(errs) > 0 && statusCode == 404 && statusCode >= 500 {
		return fmt.Errorf("cannot reach %s, status code: %d, errors: %v", body.Url, statusCode, errs)
	}

	body.Shorty = pkg.HumanFriendlyEnglishString(8)

	if err := pkg.Redis.Set(ctx.Context(), body.Shorty, body.Url, body.Expired, true); err != nil {
		return err
	}

	return ctx.JSON(types.Response{
		Error:   false,
		Message: fmt.Sprintf("%s/%s", ctx.BaseURL(), body.Shorty),
	})
}
