package model

// PullRequest represents a PR from the list endpoint.
// Note: the list endpoint does NOT return additions/deletions/changed_files.
type PullRequest struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	User      string `json:"user"`
	AvatarURL string `json:"avatar_url"`
	Branch    string `json:"branch"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Draft     bool   `json:"draft"`
}
