package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

type AuthConfig struct {
	ClientID     string
	ClientSecret string
	TokenFile    string
}

type Auth struct {
	config    *oauth2.Config
	tokenFile string
}

func NewAuth(cfg AuthConfig) *Auth {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{drive.DriveMetadataReadonlyScope},
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8085/oauth/callback",
	}

	return &Auth{
		config:    oauthCfg,
		tokenFile: cfg.TokenFile,
	}
}

func (a *Auth) GetAuthURL() string {
	return a.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func (a *Auth) ConnectWithCallback(ctx context.Context) (string, error) {
	resultCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}

	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `<html><body style="font-family:sans-serif;text-align:center;padding:60px"><h2>Authorization Failed</h2><p>No authorization code received.</p><p>You can close this tab.</p></body></html>`)
			resultCh <- fmt.Errorf("no authorization code received")
			return
		}

		err := a.Exchange(ctx, code)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `<html><body style="font-family:sans-serif;text-align:center;padding:60px"><h2>Authorization Failed</h2><p>Failed to exchange code. Please try again.</p><p>You can close this tab.</p></body></html>`)
			resultCh <- err
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body style="font-family:sans-serif;text-align:center;padding:60px"><h2>Connected!</h2><p>Google Drive connected successfully.</p><p>You can close this tab and return to the app.</p></body></html>`)
		resultCh <- nil
	})

	listener, err := net.Listen("tcp", ":8085")
	if err != nil {
		return "", fmt.Errorf("failed to start OAuth callback server: %w", err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			slog.Error("OAuth callback server error", "error", err)
		}
	}()

	go func() {
		select {
		case <-resultCh:
			time.Sleep(500 * time.Millisecond)
			server.Close()
		case <-ctx.Done():
			server.Close()
		}
	}()

	return a.GetAuthURL(), nil
}

func (a *Auth) Exchange(ctx context.Context, code string) error {
	token, err := a.config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange authorization code: %w", err)
	}
	return a.saveToken(token)
}

func (a *Auth) GetClient(ctx context.Context) (*http.Client, error) {
	token, err := a.loadToken()
	if err != nil {
		return nil, fmt.Errorf("not connected to Google Drive: %w", err)
	}

	tokenSource := a.config.TokenSource(ctx, token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if newToken.AccessToken != token.AccessToken {
		if err := a.saveToken(newToken); err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %w", err)
		}
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

func (a *Auth) IsConnected() bool {
	_, err := a.loadToken()
	return err == nil
}

func (a *Auth) Disconnect() error {
	if err := os.Remove(a.tokenFile); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove token file: %w", err)
	}
	return nil
}

func (a *Auth) saveToken(token *oauth2.Token) error {
	f, err := os.Create(a.tokenFile)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to write token: %w", err)
	}

	return nil
}

func (a *Auth) loadToken() (*oauth2.Token, error) {
	f, err := os.Open(a.tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open token file: %w", err)
	}
	defer f.Close()

	var token oauth2.Token
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}
