package domain

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
)

type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Code + ": " + e.Message
}

var (
	ErrTeamExists = Error{
		Code:    "TEAM_EXISTS",
		Message: "team_name already exists",
	}

	ErrPRExists = Error{
		Code:    "PR_EXISTS",
		Message: "PR id already exists",
	}

	ErrPRMerged = Error{
		Code:    "PR_MERGED",
		Message: "cannot reassign on merged PR",
	}

	ErrNotFound = Error{
		Code:    "NOT_FOUND",
		Message: "resource not found",
	}

	ErrBadRequest = Error{
		Code:    "BAD_REQUEST",
		Message: "bad request",
	}

	ErrInternal = Error{
		Code:    "INTERNAL_ERROR",
		Message: "internal server error",
	}
)

func NewErrorResponse(ctx context.Context, w http.ResponseWriter, e Error, statusCode int) {
	errRes := ErrorResponse{
		Error: e}

	jsonErrRes, err := json.Marshal(errRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(ContentType, ApplicationJSON)
	w.WriteHeader(statusCode)
	w.Write(jsonErrRes)
}
