package components

import "github.com/maxence-charriere/go-app/v10/pkg/app"

func Confirm(title string) bool {
	result := make(chan bool)

	app.Window().Get("Swal").Call("fire", map[string]interface{}{
		"title":              title,
		"icon":               "warning",
		"showCancelButton":   true,
		"confirmButtonText":  "Yes",
		"cancelButtonText":   "Cancel",
		"confirmButtonColor": "#3085d6",
		"cancelButtonColor":  "#d33",
	}).Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		result <- args[0].Get("isConfirmed").Bool()
		return nil
	}))

	return <-result
}

func Prompt(title string, defaultValue string) string {
	result := make(chan string)

	app.Window().Get("Swal").Call("fire", map[string]interface{}{
		"title":              title,
		"input":              "text",
		"inputValue":         defaultValue,
		"showCancelButton":   true,
		"confirmButtonColor": "#3085d6",
		"cancelButtonColor":  "#d33",
	}).Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		result <- ""
		if args[0].Get("isConfirmed").Bool() {
			result <- args[0].Get("value").String()
		}

		return nil
	}))

	return <-result
}

func ShowToast(title string, message string, icon string) {
	app.Window().Get("Swal").Call("fire", map[string]interface{}{
		"title":             title,
		"text":              message,
		"icon":              icon,
		"toast":             true,
		"position":          "top-end",
		"showConfirmButton": false,
		"timer":             3000,
		"timerProgressBar":  true,
	})
}
