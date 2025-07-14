package components

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"wasm/types"

	"github.com/goccy/go-json"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type FileUpload struct {
	app.Compo
	uploading bool
	checking  bool
	progress  int
	error     string
}

func (f *FileUpload) handleFileSelect(ctx app.Context, e app.Event) {
	files := e.Get("target").Get("files")
	if files.Length() == 0 {
		return
	}

	file := files.Index(0)
	if file.Get("size").Int() > 100*1000*1000 { // 100MB
		f.handleError(ctx, "File size exceeds limit")
		return
	}

	f.checkAndUploadFile(ctx, file)
}

func (f *FileUpload) checkAndUploadFile(ctx app.Context, file app.Value) {
	f.checking = true
	f.error = ""
	ctx.Update()

	// Create FormData for filename check
	formData := new(bytes.Buffer)
	writer := multipart.NewWriter(formData)
	err := writer.WriteField("filename", file.Get("name").String())
	if err != nil {
		f.handleError(ctx, "Failed to prepare request: "+err.Error())
		return
	}
	writer.Close()

	// Make the check request
	req, err := http.NewRequest("POST", types.API_BASE_URL+"/check-filename", formData)
	if err != nil {
		f.handleError(ctx, "Failed to create request: "+err.Error())
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := types.DefaultClient.Do(req)
	if err != nil {
		f.handleError(ctx, "Failed to check filename: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var result types.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		f.handleError(ctx, "Failed to parse response: "+err.Error())
		return
	}

	if result.Error {
		f.handleError(ctx, result.Message)
		return
	}

	// If we get here, filename is available, proceed with upload
	f.uploadFile(ctx, file)
}

func (f *FileUpload) uploadFile(ctx app.Context, file app.Value) {
	f.uploading = true
	f.progress = 0
	ctx.Update()

	// Create FormData
	formData := app.Window().Get("FormData").New()
	formData.Call("append", "file", file)

	// Create XMLHttpRequest
	xhr := app.Window().Get("XMLHttpRequest").New()
	xhr.Call("open", "POST", types.API_BASE_URL+"/upload")

	// Declare handlers
	var progressHandler, loadHandler, errorHandler app.Func

	// Define progress handler
	progressHandler = app.FuncOf(func(this app.Value, args []app.Value) any {
		if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
			if args[0].Get("lengthComputable").Bool() {
				loaded := args[0].Get("loaded").Float()
				total := args[0].Get("total").Float()
				if total > 0 {
					f.progress = int(loaded / total * 100)
					ctx.Dispatch(func(ctx app.Context) {
						ctx.Update()
					})
				}
			}
		}
		return nil
	})

	// Define load handler
	loadHandler = app.FuncOf(func(this app.Value, args []app.Value) any {
		status := xhr.Get("status").Int()
		if status == 200 {
			ctx.Dispatch(func(ctx app.Context) {
				f.reset(ctx)
				ShowToast("Success", "File uploaded successfully", "success")
			})
		} else {
			errorMsg := fmt.Sprintf("Upload failed with status %d", status)
			ctx.Dispatch(func(ctx app.Context) {
				f.handleError(ctx, errorMsg)
			})
		}
		// Clean up handlers
		if progressHandler != nil {
			progressHandler.Release()
		}
		if loadHandler != nil {
			loadHandler.Release()
		}
		if errorHandler != nil {
			errorHandler.Release()
		}
		return nil
	})

	// Define error handler
	errorHandler = app.FuncOf(func(this app.Value, args []app.Value) any {
		ctx.Dispatch(func(ctx app.Context) {
			f.handleError(ctx, "Network error occurred during upload")
		})
		// Clean up handlers
		if progressHandler != nil {
			progressHandler.Release()
		}
		if loadHandler != nil {
			loadHandler.Release()
		}
		if errorHandler != nil {
			errorHandler.Release()
		}
		return nil
	})

	// Set up event handlers
	xhr.Get("upload").Set("onprogress", progressHandler)
	xhr.Set("onload", loadHandler)
	xhr.Set("onerror", errorHandler)

	// Send the request
	xhr.Call("send", formData)
}

func (f *FileUpload) handleError(ctx app.Context, msg string) {
	f.error = msg
	f.reset(ctx)
	ShowToast("Error", msg, "error")
}

func (f *FileUpload) reset(ctx app.Context) {
	f.uploading = false
	f.checking = false
	f.progress = 0
	ctx.Update()
}

func (f *FileUpload) Render() app.UI {
	return app.Div().
		Class("rounded-lg bg-white p-6 shadow-md").
		Body(
			app.H2().
				Class("mb-4 text-xl font-semibold").
				Text("Upload File"),

			app.P().
				Class("mb-4 text-sm text-gray-600").
				Text("Maximum file size: 100MB"),

			app.If(f.error != "",
				func() app.UI {
					return app.Div().
						Class("mb-4 rounded bg-red-100 p-3 text-red-700").
						Text(f.error)
				},
			),

			app.Div().
				Class("mb-4").
				Body(
					app.Input().
						Type("file").
						Class("block w-full text-sm text-gray-500 file:mr-4 file:cursor-pointer file:rounded-full file:border-0 file:bg-blue-50 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-blue-700 file:transition-colors file:duration-200 hover:cursor-pointer hover:file:bg-blue-100").
						OnChange(f.handleFileSelect).
						Disabled(f.uploading || f.checking),
				),

			app.If(f.checking,
				func() app.UI {
					return app.P().
						Class("text-sm text-gray-600").
						Text("Checking filename availability...")
				},
			),

			app.If(f.uploading,
				func() app.UI {
					return app.Div().
						Class("mb-4").
						Body(
							app.Div().
								Class("h-2 w-full rounded-full bg-gray-200").
								Body(
									app.Div().
										Class("h-2 rounded-full bg-blue-600 transition-all duration-300").
										Style("width", fmt.Sprintf("%d%%", f.progress)),
								),
							app.P().
								Class("mt-1 text-sm text-gray-600").
								Text(fmt.Sprintf("%d%% uploaded", f.progress)),
						)
				},
			),
		)
}
