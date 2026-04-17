package api

import (
	"io/fs"
	"net/http"
)

// NewRouter sets up all routes and returns the top-level handler.
func NewRouter(h *Handler, oauth OAuthConfig, fallbackToken string, webFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// Auth routes (no middleware needed)
	mux.HandleFunc("GET /auth/login", oauth.HandleLogin)
	mux.HandleFunc("GET /auth/callback", oauth.HandleCallback)
	mux.HandleFunc("GET /auth/logout", oauth.HandleLogout)

	// API routes
	mux.HandleFunc("GET /api/user", h.GetUser)
	mux.HandleFunc("GET /api/user/repos", h.ListUserRepos)
	mux.HandleFunc("GET /api/strategies", h.ListStrategies)
	mux.HandleFunc("GET /api/repos/{owner}/{repo}/pulls", h.ListPRs)
	mux.HandleFunc("GET /api/repos/{owner}/{repo}/pulls/{number}/files", h.GetPRFiles)
	mux.HandleFunc("GET /api/raw", h.GetRawFile)

	// Serve frontend
	mux.Handle("GET /", http.FileServerFS(webFS))

	return AuthMiddleware(fallbackToken, mux)
}
