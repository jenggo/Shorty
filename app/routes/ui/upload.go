package ui

import (
	"fmt"
	"net/url"
	"runtime"
	"shorty/config"
	"shorty/pkg"
	"shorty/types"
	"shorty/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

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
			// Client disconnected/cancelled - clean up
			log.Info().Msg("upload cancelled")
			// Clean up any partial uploads
			slugifiedName := utils.SlugifyFilename(ctx.FormValue("file"))
			if err := utils.Storage.Delete(slugifiedName); err != nil {
				log.Warn().Err(err).Msg("failed to cleanup cancelled upload")
			}
		case <-done:
			// Normal completion - do nothing
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
	select {
	case <-ctx.Context().Done():
		return ctx.Status(fiber.StatusRequestTimeout).JSON(types.Response{
			Error:   true,
			Message: "Upload cancelled",
		})
	default:
		if err := ctx.SaveFileToStorage(file, slugifiedName, utils.Storage); err != nil {
			log.Error().Caller().Err(err).Send()
			return fmt.Errorf("failed save file to storage: %v", err)
		}
	}

	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "inline")
	url, err := utils.Storage.Conn().PresignedGetObject(ctx.Context(), config.Use.S3.Bucket, slugifiedName, config.Use.S3.Expired, reqParams)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return fmt.Errorf("failed to get presigned url: %v", err)
	}

	shorty := utils.HumanFriendlyEnglishString(8)
	if err := pkg.Redis.Set(ctx.Context(), shorty, url.String(), config.Use.S3.Expired, true); err != nil {
		log.Error().Caller().Err(err).Send()
		return fmt.Errorf("failed to set redis key: %v", err)
	}

	// Aggresively freeing memory
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
