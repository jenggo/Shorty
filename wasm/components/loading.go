package components

import "github.com/maxence-charriere/go-app/v10/pkg/app"

func Loading(size string) app.UI {
	return app.Div().
		Class("flex items-center justify-center").
		Body(
			app.Div().
				Class(size + " animate-spin rounded-full border-4 border-gray-200 border-t-blue-500"),
		)
}
