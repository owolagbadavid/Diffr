package api

import (
	"net/http"
)

// NewRouter sets up all routes and returns the top-level handler.
func NewRouter(h *Handler, oauth OAuthConfig, fallbackToken string) http.Handler {
	mux := http.NewServeMux()

	// Auth routes
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

	return corsMiddleware(AuthMiddleware(fallbackToken, mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
