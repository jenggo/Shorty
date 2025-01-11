package routes

import (
	"context"
	"time"

	"shorty/pkg"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Upgrade(ctx *fiber.Ctx) error {
	sess, err := sessionStore.Get(ctx)
	if err != nil {
		return err
	}

	if name := sess.Get("name"); name == nil {
		return fiber.ErrForbidden
	}

	if !websocket.IsWebSocketUpgrade(ctx) {
		log.Error().Bool("ws connection", true).Msgf("cannot upgrade to websocket for ip %s", ctx.IP())
		return fiber.ErrUpgradeRequired
	}

	return ctx.Next()
}

func Websocket(c *websocket.Conn) {
	c.EnableWriteCompression(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send initial data
	if err := sendData(c, ctx); err != nil {
		log.Error().Caller().Err(err).Msg("failed to send initial data")
		return
	}

	// Start periodic updates
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sendData(c, ctx); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func sendData(c *websocket.Conn, ctx context.Context) error {
	lists, err := pkg.Redis.GetAll(ctx)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(lists)
	if err != nil {
		return err
	}

	return c.WriteMessage(websocket.TextMessage, jsonData)
}
