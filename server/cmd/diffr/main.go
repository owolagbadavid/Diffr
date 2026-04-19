package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"diffr/internal/api"
	ghclient "diffr/internal/github"
	_ "diffr/internal/strategy"

	github_ratelimit "github.com/gofri/go-github-ratelimit/v2/github_ratelimit"
)

func main() {
	port := flag.Int("port", 3000, "HTTP server port")
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "fallback GitHub token for unauthenticated requests")
	clientID := flag.String("client-id", os.Getenv("GITHUB_CLIENT_ID"), "GitHub OAuth App client ID")
	clientSecret := flag.String("client-secret", os.Getenv("GITHUB_CLIENT_SECRET"), "GitHub OAuth App client secret")
	baseURL := flag.String("base-url", os.Getenv("BASE_URL"), "public base URL (default: http://localhost:<port>)")
	frontendURL := flag.String("frontend-url", os.Getenv("FRONTEND_URL"), "frontend URL to redirect after auth")
	flag.Parse()

	if *baseURL == "" {
		*baseURL = fmt.Sprintf("http://localhost:%d", *port)
	}

	oauth := api.OAuthConfig{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		BaseURL:      *baseURL,
		FrontendURL:  *frontendURL,
	}

	// Rate-limited HTTP client shared across all requests
	rateLimitedHTTPClient := github_ratelimit.NewClient(nil)
	baseGH := ghclient.NewClient(*token, rateLimitedHTTPClient)

	handler := api.NewHandler(baseGH)
	router := api.NewRouter(handler, oauth, *token)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("diffr running at http://localhost%s", addr)
	if *clientID != "" {
		log.Printf("GitHub OAuth enabled (client_id=%s...)", (*clientID)[:min(8, len(*clientID))])
	} else {
		log.Printf("No GITHUB_CLIENT_ID set — OAuth login disabled, using fallback token")
	}
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
