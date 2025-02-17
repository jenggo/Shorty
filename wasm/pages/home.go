package pages

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"wasm/components"
	"wasm/types"

	"github.com/goccy/go-json"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Home struct {
	app.Compo
	Auth           *components.AuthStore
	Data           []types.ShortyData
	Loading        bool
	Error          string
	Connected      bool
	ShowCreateForm bool
	ShowUploadForm bool
	NewURL         string
	CustomName     string
	FormLoading    bool
	SSE            *components.SSEHandler
}

func (h *Home) OnMount(ctx app.Context) {
	if h.Auth != nil {
		if err := h.Auth.CheckSession(); err != nil && !h.Auth.Data.IsAuthenticated {
			app.Window().Get("location").Set("href", "/web/login")
			return
		}
	}

	if h.SSE != nil {
		h.SSE.Close()
	}

	h.initializeSSE(ctx)
}

func (h *Home) OnDismount() {
	if h.SSE != nil {
		h.SSE.Close()
	}
}

func (h *Home) initializeSSE(ctx app.Context) {
	h.Loading = true

	h.SSE = components.NewSSEHandler(components.SSEConfig{
		URL:            types.API_BASE_URL + "/events",
		MaxRetries:     5,
		ReconnectDelay: 5 * time.Second,
	})

	// Handle data updates
	h.SSE.AddEventListener("message", func(data string) {
		h.Loading = false
		if err := json.Unmarshal([]byte(data), &h.Data); err != nil {
			h.Error = "Failed to parse data: " + err.Error()
		}
		ctx.Dispatch(func(ctx app.Context) {
			ctx.Update()
		})
	})

	// Mount the SSE handler
	h.SSE.Mount(ctx)

	// Update connection status when it changes
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			connected := h.SSE.IsConnected()
			if connected != h.Connected {
				h.Connected = connected
				h.Error = ""
				if !connected {
					h.Error = "Connection lost"
				}
				ctx.Dispatch(func(ctx app.Context) {
					ctx.Update()
				})
			}
		}
	}()
}

func (h *Home) handleCreate(ctx app.Context, e app.Event) {
	e.PreventDefault()
	h.FormLoading = true

	go func() {
		payload := map[string]interface{}{
			"url": h.NewURL,
		}
		if h.CustomName != "" {
			payload["shorty"] = h.CustomName
		}

		jsonData, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", types.API_BASE_URL+"/shorty", strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := types.DefaultClient.Do(req)
		if err != nil {
			h.handleError(ctx, "Failed to create shorty: "+err.Error())
			return
		}
		defer resp.Body.Close()

		var result types.APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			h.handleError(ctx, "Failed to parse response")
			return
		}

		if result.Error {
			h.handleError(ctx, result.Message)
			return
		}

		// Success
		h.FormLoading = false
		h.ShowCreateForm = false
		h.NewURL = ""
		h.CustomName = ""
		components.ShowToast("Success", "Shorty created successfully", "success")
		ctx.Dispatch(func(ctx app.Context) {
			ctx.Update()
		})
	}()
}

func (h *Home) handleDelete(ctx app.Context, shorty string) {
	if !components.Confirm("Are you sure you want to delete this shorty?") {
		return
	}

	go func() {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", types.API_BASE_URL, shorty), nil)
		resp, err := types.DefaultClient.Do(req)
		if err != nil {
			h.handleError(ctx, "Failed to delete shorty: "+err.Error())
			return
		}
		defer resp.Body.Close()

		var result types.APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			h.handleError(ctx, "Failed to parse response")
			return
		}

		if result.Error {
			h.handleError(ctx, result.Message)
			return
		}

		components.ShowToast("Success", "Shorty deleted successfully", "success")
	}()
}

func (h *Home) handleRename(ctx app.Context, oldName string) {
	newName := components.Prompt("Enter new name:", oldName)
	if newName == "" || newName == oldName {
		return
	}

	go func() {
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/%s/%s", types.API_BASE_URL, oldName, newName), nil)
		resp, err := types.DefaultClient.Do(req)
		if err != nil {
			h.handleError(ctx, "Failed to rename shorty: "+err.Error())
			return
		}
		defer resp.Body.Close()

		var result types.APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			h.handleError(ctx, "Failed to parse response")
			return
		}

		if result.Error {
			h.handleError(ctx, result.Message)
			return
		}

		components.ShowToast("Success", "Shorty renamed successfully", "success")
	}()
}

func (h *Home) handleError(ctx app.Context, msg string) {
	h.FormLoading = false
	h.Error = msg
	components.ShowToast("Error", msg, "error")
	ctx.Dispatch(func(ctx app.Context) {
		ctx.Update()
	})
}

func (h *Home) copyToClipboard(url string) {
	app.Window().Get("navigator").Get("clipboard").Call("writeText", url)
	components.ShowToast("Success", "Copied to clipboard!", "success")
}

func (h *Home) formatExpiry(duration time.Duration) string {
	secs := duration.Seconds()

	if secs < 60 {
		return fmt.Sprintf("%ds", int(secs))
	}
	if secs < 3600 {
		return fmt.Sprintf("%dm", int(secs/60))
	}
	if secs < 86400 {
		return fmt.Sprintf("%dh", int(secs/3600))
	}
	return fmt.Sprintf("%dd", int(secs/86400))
}

func (h *Home) Render() app.UI {
	return app.Div().Body(
		// Navigation Bar
		app.Nav().
			Class("bg-gray-800 p-4").
			Body(
				app.Div().
					Class("container mx-auto flex items-center justify-between").
					Body(
						app.H1().
							Class("text-xl font-bold text-white").
							Text("Hello "+h.Auth.Data.Username),
						app.Div().
							Class("flex gap-4").
							Body(
								app.Button().
									Class("rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700").
									Text(func() string {
										if h.ShowCreateForm {
											return "Close"
										}
										return "Create New"
									}()).
									OnClick(func(ctx app.Context, e app.Event) {
										h.ShowCreateForm = !h.ShowCreateForm
										ctx.Update()
									}),
								app.If(h.Auth.Data.S3Enabled,
									func() app.UI {
										return app.Button().
											Class("rounded bg-green-600 px-4 py-2 text-white hover:bg-green-700").
											Text(func() string {
												if h.ShowUploadForm {
													return "Close"
												}
												return "Upload File"
											}()).
											OnClick(func(ctx app.Context, e app.Event) {
												h.ShowUploadForm = !h.ShowUploadForm
												ctx.Update()
											})
									},
								),
								app.Button().
									Class("rounded bg-red-600 px-4 py-2 text-white hover:bg-red-700").
									Text("Logout").
									OnClick(func(ctx app.Context, e app.Event) {
										app.Window().Get("location").Set("href", "/web/logout")
									}),
							),
					),
			),

		// Main Content
		app.Main().
			Class("container mx-auto p-4").
			Body(
				// Error/Connection Status
				app.If(!h.Connected || h.Error != "",
					func() app.UI {
						return app.Div().
							Class("mb-4 rounded bg-red-100 p-4 text-red-700").
							Body(
								app.If(!h.Connected,
									func() app.UI {
										return app.Span().Text("Connection lost - attempting to reconnect...")
									},
								).Else(
									func() app.UI {
										return app.Text(h.Error)
									},
								),
								app.Button().
									Class("ml-2 text-red-500 hover:text-red-700").
									Text("Retry").
									OnClick(func(ctx app.Context, e app.Event) {
										h.SSE.Close()
										h.initializeSSE(ctx)
									}),
							)
					},
				),

				// Create Form
				app.If(h.ShowCreateForm,
					func() app.UI {
						return h.renderCreateForm()
					},
				),

				// Upload Form
				app.If(h.ShowUploadForm && h.Auth.Data.S3Enabled,
					func() app.UI {
						return app.Div().
							Class("mb-6").
							Body(
								&components.FileUpload{},
							)
					},
				),

				// Data Table
				app.If(h.Loading,
					func() app.UI {
						return app.Div().
							Class("flex justify-center p-8").
							Body(
								components.Loading("w-8 h-8"),
							)
					},
				).Else(
					func() app.UI {
						return h.renderDataTable()
					},
				),
			),
	)
}

func (h *Home) renderCreateForm() app.UI {
	return app.Div().
		Class("mb-6 rounded-lg bg-white p-4 shadow").
		Body(
			app.Form().
				Class("space-y-4").
				OnSubmit(h.handleCreate).
				Body(
					app.Div().Body(
						app.Label().
							Class("block text-sm font-medium text-gray-700").
							For("url").
							Text("URL"),
						app.Input().
							Type("url").
							ID("url").
							Class("mt-1 block w-full rounded-md border-gray-300 shadow-sm").
							Placeholder("https://example.com").
							Required(true).
							Value(h.NewURL).
							OnInput(func(ctx app.Context, e app.Event) {
								h.NewURL = e.Get("target").Get("value").String()
								ctx.Update()
							}),
					),
					app.Div().Body(
						app.Label().
							Class("block text-sm font-medium text-gray-700").
							For("customName").
							Text("Custom Name (Optional)"),
						app.Input().
							Type("text").
							ID("customName").
							Class("mt-1 block w-full rounded-md border-gray-300 shadow-sm").
							Placeholder("my-custom-url").
							Value(h.CustomName).
							OnInput(func(ctx app.Context, e app.Event) {
								h.CustomName = e.Get("target").Get("value").String()
								ctx.Update()
							}),
					),
					app.Div().
						Class("flex justify-end gap-2").
						Body(
							app.Button().
								Type("button").
								Class("rounded border px-4 py-2 text-gray-700 hover:bg-gray-50").
								Text("Cancel").
								OnClick(func(ctx app.Context, e app.Event) {
									h.ShowCreateForm = false
									ctx.Update()
								}),
							app.Button().
								Type("submit").
								Class("rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50").
								Disabled(h.FormLoading).
								Body(
									app.If(h.FormLoading,
										func() app.UI {
											return components.Loading("w-5 h-5")
										},
									).Else(
										func() app.UI {
											return app.Text("Create")
										},
									),
								),
						),
				),
		)
}

func (h *Home) renderDataTable() app.UI {
	return app.Div().
		Class("overflow-x-auto").
		Body(
			app.Table().
				Class("min-w-full divide-y divide-gray-200").
				Body(
					// Table Header
					app.THead().
						Class("bg-gray-50").
						Body(
							app.Tr().Body(
								app.Th().
									Class("px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500").
									Text("Shorty"),
								app.Th().
									Class("px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500").
									Text("File"),
								app.Th().
									Class("px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500").
									Text("URL"),
								app.Th().
									Class("px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500").
									Text("Expired"),
								app.Th().
									Class("px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500").
									Text("Actions"),
							),
						),
					// Table Body
					app.TBody().
						Class("divide-y divide-gray-200 bg-white").
						Body(
							app.Range(h.Data).Slice(func(i int) app.UI {
								row := h.Data[i]
								return app.Tr().Body(
									// Shorty column with copy button
									app.Td().
										Class("whitespace-nowrap px-6 py-4").
										Body(
											app.Button().
												Class("flex items-center gap-2 text-blue-600 hover:text-blue-800").
												OnClick(func(ctx app.Context, e app.Event) {
													h.copyToClipboard(fmt.Sprintf("%s/%s", types.API_BASE_URL, row.Shorty))
												}).
												Body(
													app.Text(row.Shorty),
													app.Raw(`<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
													</svg>`),
												),
										),
									// File column
									app.Td().
										Class("whitespace-nowrap px-6 py-4").
										Text(row.File),
									// URL column
									app.Td().
										Class("max-w-xs px-6 py-4").
										Body(
											app.A().
												Href(row.URL).
												Class("block truncate text-blue-500 hover:underline").
												Target("_blank").
												Title(row.URL).
												Text(row.URL),
										),
									// Expired column
									app.Td().
										Class("whitespace-nowrap px-6 py-4").
										Text(h.formatExpiry(row.Expired)),
									// Actions column
									app.Td().
										Class("whitespace-nowrap px-6 py-4").
										Body(
											app.Div().
												Class("flex gap-2").
												Body(
													app.Button().
														Class("rounded-md bg-blue-600 px-3 py-1 text-sm font-medium text-white hover:bg-blue-700").
														Text("Rename").
														OnClick(func(ctx app.Context, e app.Event) {
															h.handleRename(ctx, row.Shorty)
														}),
													app.Button().
														Class("rounded-md bg-red-600 px-3 py-1 text-sm font-medium text-white hover:bg-red-700").
														Text("Delete").
														OnClick(func(ctx app.Context, e app.Event) {
															h.handleDelete(ctx, row.Shorty)
														}),
												),
										),
								)
							}),
						),
				),
		)
}
