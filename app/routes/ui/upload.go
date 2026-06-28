package ui

import (
	"fmt"
	"mime/multipart"
	"net/url"
	"runtime"
	"shorty/config"
	"shorty/pkg"
	"shorty/types"
	"shorty/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func saveFileAndGetPresignedURL(ctx fiber.Ctx, file *multipart.FileHeader, slugifiedName string) (string, error) {
	select {
	case <-ctx.Context().Done():
		return "", fmt.Errorf("upload cancelled")
	default:
		if err := ctx.SaveFileToStorage(file, slugifiedName, utils.Storage); err != nil {
			return "", fmt.Errorf("failed save file to storage: %w", err)
		}
	}

	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "inline")
	presignedURL, err := utils.Storage.Conn().PresignedGetObject(ctx.Context(), config.Use.S3.Bucket, slugifiedName, config.Use.S3.Expired, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}

	return presignedURL.String(), nil
}

func Upload(ctx fiber.Ctx) error {
	if _, err := validateSession(ctx); err != nil {
		log.Error().Caller().Err(err).Send()
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Context().Done():
			log.Info().Msg("upload cancelled")
			slugifiedName := utils.SlugifyFilename(ctx.FormValue("file"))
			if err := utils.Storage.Delete(slugifiedName); err != nil {
				log.Warn().Err(err).Msg("failed to cleanup cancelled upload")
			}
		case <-done:
			return
		}
	}()

	file, err := ctx.FormFile("file")
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "Invalid file upload: " + err.Error(),
		})
	}

	slugifiedName := utils.SlugifyFilename(file.Filename)
	presignedURL, err := saveFileAndGetPresignedURL(ctx, file, slugifiedName)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return err
	}

	shorty := utils.HumanFriendlyEnglishString(8)
	if err := pkg.Redis.Set(ctx.Context(), shorty, presignedURL, config.Use.S3.Expired, true); err != nil {
		log.Error().Caller().Err(err).Send()
		return fmt.Errorf("failed to set redis key: %w", err)
	}

	runtime.GC()

	return ctx.JSON(types.Response{
		Error:   false,
		Message: fmt.Sprintf("%s/%s", ctx.BaseURL(), shorty),
	})
}

func CheckFilename(ctx fiber.Ctx) error {
	if _, err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	fileName := ctx.FormValue("filename")
	if fileName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "Filename is required",
		})
	}

	slugifiedName := utils.SlugifyFilename(fileName)
	if _, err := utils.Storage.Get(slugifiedName); err == nil {
		return ctx.Status(fiber.StatusConflict).JSON(types.Response{
			Error:   true,
			Message: fmt.Sprintf("%s already exists", slugifiedName),
		})
	}

	return ctx.JSON(types.Response{
		Error:   false,
		Message: "Filename is available",
	})
}
