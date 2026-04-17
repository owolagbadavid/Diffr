package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"deniro/internal/github"
	"deniro/internal/strategy"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ghClient creates a GitHub client from the request's auth context.
func ghClient(r *http.Request) *github.Client {
	return github.NewClient(TokenFromContext(r.Context()))
}

// GET /api/user — returns the logged-in user, or 401.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	token := TokenFromContext(r.Context())
	if token == "" {
		writeJSON(w, http.StatusOK, map[string]any{"logged_in": false})
		return
	}
	user, err := ghClient(r).GetUser()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"logged_in": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"logged_in":  true,
		"login":      user.Login,
		"name":       user.Name,
		"avatar_url": user.AvatarURL,
	})
}

// GET /api/user/repos
func (h *Handler) ListUserRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := ghClient(r).ListUserRepos()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, repos)
}

// GET /api/strategies
func (h *Handler) ListStrategies(w http.ResponseWriter, r *http.Request) {
	type strat struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	var out []strat
	for _, name := range strategy.Names() {
		s, _ := strategy.Get(name)
		out = append(out, strat{Name: s.Name(), Description: s.Description()})
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /api/repos/{owner}/{repo}/pulls
func (h *Handler) ListPRs(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	prs, err := ghClient(r).ListPRs(owner, repo)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, prs)
}

// GET /api/repos/{owner}/{repo}/pulls/{number}/files?strategy=by-size
func (h *Handler) GetPRFiles(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")
	number, err := strconv.Atoi(r.PathValue("number"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid PR number"})
		return
	}

	stratName := r.URL.Query().Get("strategy")
	if stratName == "" {
		stratName = "by-size"
	}
	s, err := strategy.Get(stratName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	files, err := ghClient(r).FetchPRFiles(owner, repo, number)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	groups := s.Organize(files)

	type response struct {
		Owner    string      `json:"owner"`
		Repo     string      `json:"repo"`
		Number   int         `json:"number"`
		Strategy string      `json:"strategy"`
		Total    int         `json:"total_files"`
		Groups   interface{} `json:"groups"`
	}
	writeJSON(w, http.StatusOK, response{
		Owner:    owner,
		Repo:     repo,
		Number:   number,
		Strategy: s.Name(),
		Total:    len(files),
		Groups:   groups,
	})
}

// GET /api/raw?url=<contents_url>
// Fetches file content via the GitHub Contents API, decodes the base64
// response, and returns plain text.
func (h *Handler) GetRawFile(w http.ResponseWriter, r *http.Request) {
	contentsURL := r.URL.Query().Get("url")
	if contentsURL == "" || !strings.HasPrefix(contentsURL, "https://api.github.com/") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing or invalid url"})
		return
	}

	token := TokenFromContext(r.Context())

	req, err := http.NewRequest("GET", contentsURL, nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "github returned " + resp.Status})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	var result struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to parse response"})
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(result.Content, "\n", ""))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to decode content"})
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(decoded)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
