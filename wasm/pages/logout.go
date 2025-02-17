package pages

import (
	"net/http"
	"wasm/components"
	"wasm/types"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Logout struct {
	app.Compo
	Auth *components.AuthStore
}

func (l *Logout) OnMount(ctx app.Context) {
	go func() {
		if _, err := http.Get(types.API_BASE_URL + "/logout"); err != nil {
			app.Log("logout error:", err)
		}

		// Reset auth store
		l.Auth.SetData(components.AuthStoreData{
			IsAuthenticated: false,
			Username:        "",
			S3Enabled:       false,
		})

		// Redirect to login page
		app.Window().Get("location").Set("href", "/web/login")
	}()
}

func (l *Logout) Render() app.UI {
	return app.Div().
		Class("flex min-h-screen items-center justify-center").
		Body(
			app.Text("Logging out..."),
		)
}
