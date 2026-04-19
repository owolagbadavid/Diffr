package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"diffr/internal/model"

	gh "github.com/google/go-github/v84/github"
)

// Client talks to the GitHub API.
type Client struct {
	Token      string
	HTTPClient *http.Client
	ghClient   *gh.Client
}

// NewClient creates a GitHub client with rate-limited HTTP and go-github.
func NewClient(token string, httpClient *http.Client) *Client {
	ghc := gh.NewClient(httpClient)
	if token != "" {
		ghc = ghc.WithAuthToken(token)
	}
	return &Client{
		Token:      token,
		HTTPClient: httpClient,
		ghClient:   ghc,
	}
}

type apiResponse struct {
	Body    []byte
	HasNext bool
	HasPrev bool
}

type ghPR struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Draft     bool   `json:"draft"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
}

// GetUser returns the authenticated user.
func (c *Client) GetUser() (*model.User, error) {
	resp, err := c.getAPI("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	var raw struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return nil, fmt.Errorf("decoding user: %w", err)
	}
	return &model.User{Login: raw.Login, Name: raw.Name, AvatarURL: raw.AvatarURL}, nil
}

// ListUserRepos returns repositories for the authenticated user.
func (c *Client) ListUserRepos(page int) ([]model.Repository, bool, bool, error) {
	apiURL := fmt.Sprintf("https://api.github.com/user/repos?sort=updated&per_page=30&page=%d&affiliation=owner,collaborator,organization_member", page)
	resp, err := c.getAPI(apiURL)
	if err != nil {
		return nil, false, false, err
	}
	var raw []struct {
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
		Language    string `json:"language"`
		Stars       int    `json:"stargazers_count"`
		OpenIssues  int    `json:"open_issues_count"`
		UpdatedAt   string `json:"updated_at"`
		Owner       struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		} `json:"owner"`
	}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return nil, false, false, fmt.Errorf("decoding repos: %w", err)
	}

	repos := make([]model.Repository, len(raw))
	for i, r := range raw {
		repos[i] = model.Repository{
			Name:        r.Name,
			FullName:    r.FullName,
			Description: r.Description,
			Private:     r.Private,
			Language:    r.Language,
			Stars:       r.Stars,
			OpenIssues:  r.OpenIssues,
			UpdatedAt:   r.UpdatedAt,
			OwnerLogin:  r.Owner.Login,
			OwnerAvatar: r.Owner.AvatarURL,
		}
	}
	return repos, resp.HasNext, resp.HasPrev, nil
}

// ListPRs returns open pull requests for a repo.
func (c *Client) ListPRs(owner, repo string, page int) ([]model.PullRequest, bool, bool, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=open&per_page=30&page=%d", owner, repo, page)

	resp, err := c.getAPI(apiURL)
	if err != nil {
		return nil, false, false, err
	}

	var raw []ghPR
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return nil, false, false, fmt.Errorf("decoding response: %w", err)
	}

	prs := make([]model.PullRequest, len(raw))
	for i, p := range raw {
		prs[i] = model.PullRequest{
			Number:    p.Number,
			Title:     p.Title,
			State:     p.State,
			User:      p.User.Login,
			AvatarURL: p.User.AvatarURL,
			Branch:    p.Head.Ref,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			Draft:     p.Draft,
		}
	}
	return prs, resp.HasNext, resp.HasPrev, nil
}

// FetchPRFiles returns all changed files for a pull request.
func (c *Client) FetchPRFiles(ctx context.Context, owner, repo string, number int) ([]model.FileDiff, error) {
	ctx = context.WithValue(ctx, gh.BypassRateLimitCheck, true)

	var files []model.FileDiff
	for f, err := range c.ghClient.PullRequests.ListFilesIter(ctx, owner, repo, number, &gh.ListOptions{PerPage: 100}) {
		if err != nil {
			return nil, fmt.Errorf("listing PR files: %w", err)
		}
		files = append(files, model.FileDiff{
			Filename:    f.GetFilename(),
			Status:      f.GetStatus(),
			Additions:   f.GetAdditions(),
			Deletions:   f.GetDeletions(),
			Patch:       f.GetPatch(),
			BlobURL:     f.GetBlobURL(),
			ContentsURL: f.GetContentsURL(),
		})
	}
	return files, nil
}

// FetchFileContent fetches file content via the GitHub Contents API.
func (c *Client) FetchFileContent(ctx context.Context, contentsURL string) (string, error) {
	owner, repo, path, ref, err := parseContentsURL(contentsURL)
	if err != nil {
		return "", err
	}

	ctx = context.WithValue(ctx, gh.BypassRateLimitCheck, true)
	fileContent, _, _, err := c.ghClient.Repositories.GetContents(ctx, owner, repo, path, &gh.RepositoryContentGetOptions{Ref: ref})
	if err != nil {
		return "", fmt.Errorf("fetching file: %w", err)
	}
	if fileContent == nil {
		return "", fmt.Errorf("path is a directory, not a file")
	}
	return fileContent.GetContent()
}

func fmtTime(t gh.Timestamp) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseContentsURL(contentsURL string) (owner, repo, path, ref string, err error) {
	u, err := url.Parse(contentsURL)
	if err != nil {
		return "", "", "", "", err
	}
	trimmed := strings.TrimPrefix(u.Path, "/repos/")
	parts := strings.SplitN(trimmed, "/", 4)
	if len(parts) < 4 || parts[2] != "contents" {
		return "", "", "", "", fmt.Errorf("invalid contents URL: %s", contentsURL)
	}
	return parts[0], parts[1], parts[3], u.Query().Get("ref"), nil
}

// SearchCode searches for code in a repository via the GitHub code search API.
// Returns the raw JSON response including text_matches fragments.
func (c *Client) SearchCode(owner, repo, query string) (json.RawMessage, error) {
	q := url.QueryEscape(fmt.Sprintf("%s repo:%s/%s", query, owner, repo))
	apiURL := fmt.Sprintf("https://api.github.com/search/code?q=%s&per_page=20", q)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	// text-match+json enables the text_matches field with match fragments
	req.Header.Set("Accept", "application/vnd.github.text-match+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned %s", resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	return json.RawMessage(buf), nil
}

// getAPI performs a GET request and parses pagination from the Link header.
func (c *Client) getAPI(apiURL string) (*apiResponse, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned %s", resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	link := resp.Header.Get("Link")
	return &apiResponse{
		Body:    buf,
		HasNext: strings.Contains(link, `rel="next"`),
		HasPrev: strings.Contains(link, `rel="prev"`),
	}, nil
}
