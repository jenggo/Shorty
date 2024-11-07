package routes

import (
	"shorty/config"
	"shorty/types"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var store = session.New()

func HTMLMain(ctx *fiber.Ctx) error {
	sess, err := store.Get(ctx)
	if err != nil {
		return err
	}

	if name := sess.Get("name"); name == nil {
		return ctx.Render("login", nil)
	}

	return ctx.Render("index", nil)
}

// JSON Response
func HTMLLogin(ctx *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.Username != config.Use.App.Auth.User || !checkPassword(body.Password) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: "Invalid credentials",
		})
	}

	sess, err := store.Get(ctx)
	if err != nil {
		return err
	}
	sess.Set("name", body.Username)
	if err := sess.Save(); err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(types.Response{
		Error:   false,
		Message: "Authorized",
	})
}

func HTMLLogout(ctx *fiber.Ctx) error {
	sess, err := store.Get(ctx)
	if err != nil {
		return err
	}

	if err := sess.Destroy(); err != nil {
		return err
	}

	return ctx.Redirect("/")
}

func checkPassword(input string) bool {
	pass := []byte(config.Use.App.Auth.Password)
	hashed := []byte(input)

	if err := bcrypt.CompareHashAndPassword(hashed, pass); err != nil {
		log.Error().Err(err).Send()
		return false
	}

	return true
}
