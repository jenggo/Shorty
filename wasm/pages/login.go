package pages

import (
	"net/http"
	"wasm/components"
	"wasm/types"

	"github.com/goccy/go-json"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Login struct {
	app.Compo
	Auth    *components.AuthStore
	loading bool
	error   string
}

func (l *Login) OnMount(ctx app.Context) {
	// If already authenticated, redirect to home
	if l.Auth != nil {
		if err := l.Auth.CheckSession(); err == nil && l.Auth.Data.IsAuthenticated {
			app.Window().Get("location").Set("href", "/web")
			return
		}
	}

	// Check URL parameters for error message
	urlParams := app.Window().URL().Query()
	if errorMsg := urlParams.Get("error"); errorMsg != "" {
		l.error = errorMsg
		components.ShowToast("Login Error", errorMsg, "error")
	}
}

func (l *Login) Render() app.UI {
	return app.Div().
		Class("flex min-h-screen items-center justify-center bg-gray-50").
		Body(
			app.Div().
				Class("w-96 rounded-lg bg-white p-8 shadow-md").
				Body(
					// Add heading
					app.H2().
						Class("mb-6 text-center text-2xl font-bold text-gray-800").
						Text("Login to Shorty"),

					// Add button container
					app.Div().
						Class("space-y-4").
						Body(
							app.Button().
								ID("button-login").
								Class("flex w-full items-center justify-center rounded-md px-4 py-2 text-white disabled:opacity-50").
								OnClick(l.HandleLogin).
								Body(
									app.If(l.loading,
										func() app.UI {
											return app.Div().
												Class("mr-2 h-5 w-5 animate-spin rounded-full border-2 border-white border-t-transparent")
										},
									).Else(
										func() app.UI {
											return app.Raw(`<svg class="mr-2 h-5 w-5" viewBox="0 0 586 559">
                                                <path fill="currentColor" d="M461.17 301.83l-18.91-58.12-37.42-115.28c-1.92-5.9-7.15-10.05-13.37-10.05s-11.45 4.15-13.37 10.05l-37.42 115.28h-126.5l-37.42-115.28c-1.92-5.9-7.15-10.05-13.37-10.05s-11.45 4.15-13.37 10.05l-37.42 115.28-18.91 58.12c-1.72 5.3.12 11.11 4.72 14.38l212.49 154.41 212.49-154.41c4.6-3.27 6.44-9.08 4.72-14.38"/>
                                            </svg>`)
										},
									),
									app.Text("Sign in with Repo Nusatek"),
								),
						),
				),
		)
}

func (l *Login) HandleLogin(ctx app.Context, e app.Event) {
	l.loading = true
	ctx.Dispatch(func(ctx app.Context) {
		ctx.Update()
	})

	go func() {
		resp, err := http.Get(types.API_BASE_URL + "/web/auth/gitlab")
		if err != nil {
			l.handleError(ctx, "Failed to initiate login: "+err.Error())
			return
		}
		defer resp.Body.Close()

		var result struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			l.handleError(ctx, "Invalid response from server")
			return
		}

		if result.Error {
			l.handleError(ctx, result.Message)
			return
		}

		if result.Message != "" {
			app.Window().Get("location").Set("href", result.Message)
		} else {
			l.handleError(ctx, "Invalid response from server")
		}
	}()
}

func (l *Login) handleError(ctx app.Context, msg string) {
	l.loading = false
	l.error = msg
	ctx.Dispatch(func(ctx app.Context) {
		ctx.Update()
	})
	components.ShowToast("Login Error", msg, "error")
}
