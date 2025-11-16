package domain

type TeamRequest struct {
	TeamName string `json:"team_name"`
	Members  []User `json:"members"`
}

type SetActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type CreatePRRequest struct {
	PRID     string `json:"pull_request_id"`
	PRName   string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type SetMergedRequest struct {
	PrID string `json:"pull_request_id"`
}

type ReassignRequest struct {
	PrID   string `json:"pull_request_id"`
	UserID string `json:"old_reviewer_id"`
}
