package routes

import (
	"errors"
	"shorty/config"
	"shorty/types"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	store                 = session.New()
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func validatePassword(hashedPassword, plainPassword string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword)); err != nil {
		log.Error().Err(err).Msg("password validation failed")
		return false
	}

	return true
}

func getSession(ctx *fiber.Ctx) (*session.Session, error) {
	sess, err := store.Get(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session")
		return nil, err
	}

	return sess, nil
}

func validateSession(ctx *fiber.Ctx) error {
	sess, err := getSession(ctx)
	if err != nil {
		return err
	}

	if sess.Get("name") == nil {
		return ErrUnauthorized
	}

	return nil
}

func UILogin(ctx *fiber.Ctx) error {
	var req loginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "invalid request body",
		})
	}

	if req.Username != config.Use.App.Auth.User ||
		!validatePassword(req.Password, config.Use.App.Auth.Password) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: ErrInvalidCredentials.Error(),
		})
	}

	sess, err := getSession(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(types.Response{
			Error:   true,
			Message: "session error",
		})
	}

	sess.Set("name", req.Username)
	if err := sess.Save(); err != nil {
		log.Error().Err(err).Msg("failed to save session")
		return ctx.Status(fiber.StatusInternalServerError).JSON(types.Response{
			Error:   true,
			Message: "failed to create session",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(types.Response{
		Error:   false,
		Message: "authorized",
	})
}

func UILogout(ctx *fiber.Ctx) error {
	sess, err := getSession(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(types.Response{
			Error:   true,
			Message: "session error",
		})
	}

	if err := sess.Destroy(); err != nil {
		log.Error().Err(err).Msg("failed to destroy session")
		return ctx.Status(fiber.StatusInternalServerError).JSON(types.Response{
			Error:   true,
			Message: "failed to logout",
		})
	}

	return ctx.Redirect("/")
}

func UICreate(ctx *fiber.Ctx) error {
	if err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	return Shorten(ctx)
}

func UIDelete(ctx *fiber.Ctx) error {
	if err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	return Delete(ctx)
}

func UIChange(ctx *fiber.Ctx) error {
	if err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	return Change(ctx)
}
