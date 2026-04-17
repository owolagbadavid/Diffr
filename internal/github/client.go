package github

import (
	"deniro/internal/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client talks to the GitHub REST API.
type Client struct {
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a GitHub client. Token can be empty for public repos.
func NewClient(token string) *Client {
	return &Client{Token: token, HTTPClient: http.DefaultClient}
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

type ghFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch"`
	BlobURL     string `json:"blob_url"`
	ContentsURL string `json:"contents_url"`
}

// GetUser returns the authenticated user.
func (c *Client) GetUser() (*model.User, error) {
	body, err := c.get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	var raw struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decoding user: %w", err)
	}
	return &model.User{Login: raw.Login, Name: raw.Name, AvatarURL: raw.AvatarURL}, nil
}

// ListUserRepos returns repositories for the authenticated user.
func (c *Client) ListUserRepos() ([]model.Repository, error) {
	body, err := c.get("https://api.github.com/user/repos?sort=updated&per_page=50&affiliation=owner,collaborator,organization_member")
	if err != nil {
		return nil, err
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
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decoding repos: %w", err)
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
	return repos, nil
}

// ListPRs returns open pull requests for a repo.
func (c *Client) ListPRs(owner, repo string) ([]model.PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=open&per_page=30", owner, repo)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var raw []ghPR
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
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
	return prs, nil
}

// FetchPRFiles returns the changed files for a pull request.
func (c *Client) FetchPRFiles(owner, repo string, number int) ([]model.FileDiff, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files?per_page=100", owner, repo, number)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var raw []ghFile
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	files := make([]model.FileDiff, len(raw))
	for i, f := range raw {
		files[i] = model.FileDiff{
			Filename:  f.Filename,
			Status:    f.Status,
			Additions: f.Additions,
			Deletions: f.Deletions,
			Patch:     f.Patch,
			BlobURL:     f.BlobURL,
			ContentsURL: f.ContentsURL,
		}
	}
	return files, nil
}

func (c *Client) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
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
	return buf, nil
}
