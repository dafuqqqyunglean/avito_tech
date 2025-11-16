package domain

import (
	"encoding/json"
	"net/http"
	"time"
)

type TeamResponse struct {
	Team TeamRequest `json:"team"`
}

func WriteResponse(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set(ContentType, ApplicationJSON)
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(data)
}

type SetActiveResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type CutPullRequest struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

type GetReviewResponse struct {
	UserID       string           `json:"user_id"`
	PullRequests []CutPullRequest `json:"pull_requests"`
}

type CreatePRResponse struct {
	PR PullRequest `json:"pr"`
}

type MergePRResponse struct {
	PR       PullRequest `json:"pr"`
	MergedAt time.Time   `json:"mergedAt"`
}

type ReassignPRResponse struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}
