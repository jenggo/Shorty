package routes

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"shorty/pkg"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Upgrade(ctx *fiber.Ctx) error {
	sess, err := store.Get(ctx)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return err
	}

	if name := sess.Get("name"); name == nil {
		return ctx.Render("login", nil)
	}

	if !websocket.IsWebSocketUpgrade(ctx) {
		log.Error().Bool("ws connection", true).Msgf("cannot upgrade to websocket for ip %s", ctx.IP())
		return fiber.ErrUpgradeRequired
	}

	return ctx.Next()
}

func Websocket(c *websocket.Conn) {
	tmpl, err := template.ParseFiles("ui/tbody.tpl")
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return
	}

	var w bytes.Buffer
	c.EnableWriteCompression(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		lists, err := pkg.Redis.GetAll(ctx)
		if err != nil {
			log.Error().Caller().Err(err).Send()
			break
		}

		if err := tmpl.Execute(&w, lists); err != nil {
			log.Error().Caller().Err(err).Send()
			break
		}

		if err := c.WriteMessage(websocket.TextMessage, w.Bytes()); err != nil {
			if err := c.Close(); err != nil {
				log.Error().Caller().Err(err).Send()
			}
			break
		}
		w.Reset()

		time.Sleep(time.Second)
	}
}
