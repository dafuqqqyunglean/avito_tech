package domain

type User struct {
	ID       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	ID      int    `json:"team_id"`
	Name    string `json:"team_name"`
	Members []User `json:"members"`
}

type PullRequest struct {
	ID        string   `json:"pull_request_id"`
	Name      string   `json:"pull_request_name"`
	AuthorID  string   `json:"author_id"`
	Status    string   `json:"status"`
	Reviewers []string `json:"assigned_reviewers"`
}
