package components

import (
	"net/http"
	"wasm/types"

	"github.com/goccy/go-json"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type AuthStoreData struct {
	IsAuthenticated bool
	Username        string
	S3Enabled       bool
}

type AuthStore struct {
	Data        AuthStoreData
	subscribers []func(AuthStoreData)
}

func NewAuthStore() *AuthStore {
	return &AuthStore{
		Data: AuthStoreData{
			IsAuthenticated: false,
			Username:        "",
			S3Enabled:       false,
		},
	}
}

func (a *AuthStore) Subscribe(callback func(AuthStoreData)) {
	a.subscribers = append(a.subscribers, callback)
}

func (a *AuthStore) SetData(data AuthStoreData) {
	a.Data = data
	for _, callback := range a.subscribers {
		callback(data)
	}
}

func (a *AuthStore) CheckSession() error {
	req, err := http.NewRequest(http.MethodGet, types.API_BASE_URL+"/auth/check", nil)
	if err != nil {
		return err
	}
	req.Header.Add("credentials", "include")

	resp, err := types.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Error bool `json:"error"`
		Data  struct {
			Username  string `json:"username"`
			S3Enabled bool   `json:"s3Enabled"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Error && result.Data.Username != "" {
		a.SetData(AuthStoreData{
			IsAuthenticated: true,
			Username:        result.Data.Username,
			S3Enabled:       result.Data.S3Enabled,
		})
		return nil
	}

	a.SetData(AuthStoreData{
		IsAuthenticated: false,
		Username:        "",
		S3Enabled:       false,
	})
	return nil
}

func (a *AuthStore) RequireAuth(next func(ctx app.Context)) func(ctx app.Context) {
	return func(ctx app.Context) {
		if !a.Data.IsAuthenticated {
			app.Window().Get("location").Set("href", "/web/login")
			return
		}
		next(ctx)
	}
}
