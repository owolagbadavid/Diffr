package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"deniro/internal/api"
	_ "deniro/internal/strategy"
	"deniro/web"
)

func main() {
	port := flag.Int("port", 3000, "HTTP server port")
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "fallback GitHub token for unauthenticated requests")
	clientID := flag.String("client-id", os.Getenv("GITHUB_CLIENT_ID"), "GitHub OAuth App client ID")
	clientSecret := flag.String("client-secret", os.Getenv("GITHUB_CLIENT_SECRET"), "GitHub OAuth App client secret")
	baseURL := flag.String("base-url", "", "public base URL (default: http://localhost:<port>)")
	flag.Parse()

	if *baseURL == "" {
		*baseURL = fmt.Sprintf("http://localhost:%d", *port)
	}

	oauth := api.OAuthConfig{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		BaseURL:      *baseURL,
	}

	handler := api.NewHandler()
	router := api.NewRouter(handler, oauth, *token, web.Assets)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("deniro running at http://localhost%s", addr)
	if *clientID != "" {
		log.Printf("GitHub OAuth enabled (client_id=%s...)", (*clientID)[:min(8, len(*clientID))])
	} else {
		log.Printf("No GITHUB_CLIENT_ID set — OAuth login disabled, using fallback token")
	}
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
