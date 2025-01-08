package routes

import (
	"fmt"
	"strings"

	"shorty/config"
	"shorty/pkg"
	"shorty/types"

	"github.com/gofiber/fiber/v2"
)

func Delete(ctx *fiber.Ctx) error {
	shorturl := ctx.Params("shorty")

	key, err := pkg.Redis.Get(ctx.Context(), shorturl)
	if err != nil {
		return err
	}

	if err := pkg.Redis.Del(ctx.Context(), shorturl); err != nil {
		return err
	}

	if config.Use.S3.Enable {
		prefix := fmt.Sprintf("https://%s/%s/", config.Use.S3.Endpoint, config.Use.S3.Bucket)
		if strings.HasPrefix(key, prefix) {
			getObjectName := strings.TrimPrefix(key, prefix)
			objectName := strings.SplitN(getObjectName, "?", 2)[0]

			s3, err := pkg.NewS3(ctx.Context())
			if err != nil {
				return err
			}

			if err := s3.Delete(ctx.Context(), objectName); err != nil {
				return err
			}
		}
	}

	return ctx.JSON(types.Response{
		Error:   false,
		Message: fmt.Sprintf("%s deleted", shorturl),
	})
}
