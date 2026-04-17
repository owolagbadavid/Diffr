package model

type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Language    string `json:"language"`
	Stars       int    `json:"stars"`
	OpenIssues  int    `json:"open_issues"`
	UpdatedAt   string `json:"updated_at"`
	OwnerLogin  string `json:"owner_login"`
	OwnerAvatar string `json:"owner_avatar"`
}
