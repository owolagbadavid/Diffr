package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type contextKey string

const tokenKey contextKey = "github_token"

// OAuthConfig holds GitHub OAuth credentials.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	BaseURL      string // e.g. "http://localhost:3000"
}

func (o OAuthConfig) redirectURI() string {
	return o.BaseURL + "/auth/callback"
}

// HandleLogin redirects the user to GitHub's OAuth authorize page.
func (o OAuthConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	u := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s",
		url.QueryEscape(o.ClientID),
		url.QueryEscape(o.redirectURI()),
		url.QueryEscape("read:user repo"),
	)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

// HandleCallback exchanges the code for a token and sets a cookie.
func (o OAuthConfig) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	data := url.Values{
		"client_id":     {o.ClientID},
		"client_secret": {o.ClientSecret},
		"code":          {code},
		"redirect_uri":  {o.redirectURI()},
	}

	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "oauth exchange failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
	}
	json.Unmarshal(body, &result)

	if result.AccessToken == "" {
		http.Error(w, "no access token: "+result.Error, http.StatusBadGateway)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "gh_token",
		Value:    result.AccessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30, // 30 days
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleLogout clears the auth cookie.
func (o OAuthConfig) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "gh_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// AuthMiddleware extracts the token from the cookie and adds it to the context.
func AuthMiddleware(fallbackToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := fallbackToken
		if c, err := r.Cookie("gh_token"); err == nil && c.Value != "" {
			token = c.Value
		}
		ctx := context.WithValue(r.Context(), tokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TokenFromContext returns the GitHub token from the request context.
func TokenFromContext(ctx context.Context) string {
	t, _ := ctx.Value(tokenKey).(string)
	return t
}
