package main

import (
	"wasm/components"
	"wasm/pages"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
	auth := components.NewAuthStore()

	// Initialize components
	app.Route("/web", func() app.Composer { return &pages.Home{Auth: auth} })
	// app.Route("/web/", func() app.Composer { return &pages.Home{Auth: auth} })
	app.Route("/web/login", func() app.Composer { return &pages.Login{Auth: auth} })
	app.Route("/web/logout", func() app.Composer { return &pages.Logout{Auth: auth} })

	go func() {
		if err := auth.CheckSession(); err != nil {
			app.Log(err)
		}
	}()

	// Run the app
	app.RunWhenOnBrowser()
}
