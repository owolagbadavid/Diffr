package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"deniro/internal/github"
	"deniro/internal/strategy"
)

type Handler struct {
	baseGH *github.Client
}

func NewHandler(baseGH *github.Client) *Handler {
	return &Handler{baseGH: baseGH}
}

// ghClient creates a per-request GitHub client with the user's token.
func (h *Handler) ghClient(r *http.Request) *github.Client {
	token := TokenFromContext(r.Context())
	return github.NewClient(token, h.baseGH.HTTPClient)
}

// GET /api/user
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	token := TokenFromContext(r.Context())
	if token == "" {
		writeJSON(w, http.StatusOK, map[string]any{"logged_in": false})
		return
	}
	user, err := h.ghClient(r).GetUser()
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
	repos, err := h.ghClient(r).ListUserRepos()
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

	prs, err := h.ghClient(r).ListPRs(owner, repo)
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

	files, err := h.ghClient(r).FetchPRFiles(r.Context(), owner, repo, number)
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
func (h *Handler) GetRawFile(w http.ResponseWriter, r *http.Request) {
	contentsURL := r.URL.Query().Get("url")
	if contentsURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing url parameter"})
		return
	}

	content, err := h.ghClient(r).FetchFileContent(r.Context(), contentsURL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
