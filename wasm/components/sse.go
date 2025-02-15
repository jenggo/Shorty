package components

import (
	"sync"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type SSEConfig struct {
	URL            string
	MaxRetries     int
	ReconnectDelay time.Duration
}

type SSEHandler struct {
	app.Compo
	url            string
	eventSource    app.Value
	handlers       map[string][]func(string)
	retryAttempts  int
	maxRetries     int
	reconnectDelay time.Duration
	isConnected    bool
	mu             sync.RWMutex
}

func NewSSEHandler(config SSEConfig) *SSEHandler {
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.ReconnectDelay == 0 {
		config.ReconnectDelay = 5 * time.Second
	}

	return &SSEHandler{
		url:            config.URL,
		handlers:       make(map[string][]func(string)),
		maxRetries:     config.MaxRetries,
		reconnectDelay: config.ReconnectDelay,
	}
}

func (h *SSEHandler) Mount(ctx app.Context) {
	h.connect(ctx)
}

func (h *SSEHandler) connect(ctx app.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.eventSource != nil && !h.eventSource.IsUndefined() {
		h.eventSource.Call("close")
	}

	eventSourceOptions := app.Window().Get("Object").New()
	eventSourceOptions.Set("withCredentials", true)

	h.eventSource = app.Window().Get("EventSource").New(h.url, eventSourceOptions)

	// Handle connection open
	h.eventSource.Call("addEventListener", "open", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		h.mu.Lock()
		h.isConnected = true
		h.retryAttempts = 0
		h.mu.Unlock()

		app.Log("SSE connection opened")
		return nil
	}))

	// Handle errors
	h.eventSource.Call("addEventListener", "error", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		app.Log("SSE connection error")
		h.mu.Lock()
		h.isConnected = false
		h.mu.Unlock()

		if h.retryAttempts < h.maxRetries {
			go h.reconnect(ctx)
		} else {
			app.Log("Max retry attempts reached")
			h.showErrorNotification()
		}
		return nil
	}))

	// Reattach existing event listeners
	for event, callbacks := range h.handlers {
		for _, callback := range callbacks {
			h.attachEventListener(event, callback)
		}
	}
}

func (h *SSEHandler) reconnect(ctx app.Context) {
	time.Sleep(h.reconnectDelay)

	h.mu.Lock()
	h.retryAttempts++
	attempt := h.retryAttempts
	h.mu.Unlock()

	app.Log("Attempting to reconnect (attempt %d/%d)", attempt, h.maxRetries)
	h.connect(ctx)
}

func (h *SSEHandler) AddEventListener(event string, callback func(string)) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handlers[event] = append(h.handlers[event], callback)
	h.attachEventListener(event, callback)
}

func (h *SSEHandler) attachEventListener(event string, callback func(string)) {
	if h.eventSource == nil || h.eventSource.IsUndefined() {
		return
	}

	h.eventSource.Call("addEventListener", event, app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		data := args[0].Get("data").String()
		callback(data)
		return nil
	}))
}

func (h *SSEHandler) showErrorNotification() {
	notification := app.Window().
		Get("document").
		Call("createElement", "div")

	notification.Set("className", "fixed top-4 right-4 bg-red-500 text-white px-6 py-3 rounded shadow-lg z-50")
	notification.Set("innerHTML", `
        <div class="flex items-center">
            <span class="mr-2">⚠️</span>
            <span>Connection lost. Please refresh the page or try again later.</span>
            <button class="ml-4 hover:text-red-200" onclick="this.parentElement.parentElement.remove()">✕</button>
        </div>
    `)

	app.Window().
		Get("document").
		Get("body").
		Call("appendChild", notification)

	go func() {
		time.Sleep(10 * time.Second)
		app.Window().Get("document").Call("removeChild", notification)
	}()
}

func (h *SSEHandler) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.eventSource != nil && !h.eventSource.IsUndefined() {
		h.eventSource.Call("close")
		h.eventSource = nil
	}
	h.handlers = make(map[string][]func(string))
	h.isConnected = false
}

func (h *SSEHandler) IsConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isConnected
}
