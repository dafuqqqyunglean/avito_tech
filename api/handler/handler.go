package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/dafuqqqyunglean/avito_tech/domain"
	prserv "github.com/dafuqqqyunglean/avito_tech/service/pr"
	teamserv "github.com/dafuqqqyunglean/avito_tech/service/team"
	userserv "github.com/dafuqqqyunglean/avito_tech/service/user"
)

func CreateTeam(ctx context.Context, service teamserv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.TeamRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed to decode create team request", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)

			return
		}

		resp, err := service.CreateTeam(ctx, req)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrTeamExists):
				domain.NewErrorResponse(ctx, w, domain.ErrTeamExists, http.StatusBadRequest)
			case errors.Is(err, domain.ErrNotFound):
				domain.NewErrorResponse(ctx, w, domain.ErrNotFound, http.StatusNotFound)
			case errors.Is(err, domain.ErrBadRequest):
				domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)
			default:
				slog.Error("failed to create team", "error", err)
				domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)
			}

			return
		}

		if err = domain.WriteResponse(w, http.StatusOK, resp); err != nil {
			slog.Error("failed to write response", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)

			return
		}
	}
}

func GetTeam(ctx context.Context, service teamserv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teamName := r.URL.Query().Get("team_name")

		resp, err := service.GetTeam(ctx, teamName)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				domain.NewErrorResponse(ctx, w, domain.ErrNotFound, http.StatusNotFound)
			case errors.Is(err, domain.ErrBadRequest):
				domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)
			default:
				slog.Error("failed to get team", "error", err)
				domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)
			}

			return
		}

		if err = domain.WriteResponse(w, http.StatusOK, resp); err != nil {
			slog.Error("failed to write response", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)

			return
		}
	}
}

func SetActive(ctx context.Context, service userserv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.SetActiveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed to decode set active request", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)

			return
		}

		resp, err := service.SetActive(ctx, req)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				domain.NewErrorResponse(ctx, w, domain.ErrNotFound, http.StatusNotFound)
			case errors.Is(err, domain.ErrBadRequest):
				domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)
			default:
				slog.Error("failed to set user active status", "error", err)
				domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)
			}

			return
		}

		if err = domain.WriteResponse(w, http.StatusOK, resp); err != nil {
			slog.Error("failed to write response", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)

			return
		}
	}
}

func GetReview(ctx context.Context, service userserv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")

		resp, err := service.GetReview(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				domain.NewErrorResponse(ctx, w, domain.ErrNotFound, http.StatusNotFound)
			case errors.Is(err, domain.ErrBadRequest):
				domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)
			default:
				slog.Error("failed to get user pull requests", "error", err)
				domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)
			}

			return
		}

		if err = domain.WriteResponse(w, http.StatusOK, resp); err != nil {
			slog.Error("failed to write response", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)

			return
		}
	}
}

func CreatePullRequest(ctx context.Context, service prserv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.CreatePRRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed to decode create pr request", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)

			return
		}

		resp, err := service.Create(ctx, req)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				domain.NewErrorResponse(ctx, w, domain.ErrNotFound, http.StatusNotFound)
			case errors.Is(err, domain.ErrBadRequest):
				domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)
			case errors.Is(err, domain.ErrPRExists):
				domain.NewErrorResponse(ctx, w, domain.ErrPRExists, http.StatusConflict)
			default:
				slog.Error("failed create pull request", "error", err)
				domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)
			}

			return
		}

		if err = domain.WriteResponse(w, http.StatusOK, resp); err != nil {
			slog.Error("failed to write response", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)

			return
		}
	}
}

func SetMerged(ctx context.Context, service prserv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.SetMergedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed to decode set merged request", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)

			return
		}

		resp, err := service.SetMerged(ctx, req.PrID)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				domain.NewErrorResponse(ctx, w, domain.ErrNotFound, http.StatusNotFound)
			case errors.Is(err, domain.ErrBadRequest):
				domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)
			default:
				slog.Error("failed set merged pull request", "error", err)
				domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)
			}

			return
		}

		if err = domain.WriteResponse(w, http.StatusOK, resp); err != nil {
			slog.Error("failed to write response", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)

			return
		}
	}
}

func Reassign(ctx context.Context, service prserv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.ReassignRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed to decode reassign request", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)

			return
		}

		resp, err := service.Reassign(ctx, req.PrID, req.UserID)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				domain.NewErrorResponse(ctx, w, domain.ErrNotFound, http.StatusNotFound)
			case errors.Is(err, domain.ErrBadRequest):
				domain.NewErrorResponse(ctx, w, domain.ErrBadRequest, http.StatusBadRequest)
			case errors.Is(err, domain.ErrPRMerged):
				domain.NewErrorResponse(ctx, w, domain.ErrPRMerged, http.StatusBadRequest)
			default:
				slog.Error("failed to reassign reviewer", "error", err)
				domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)
			}

			return
		}

		if err = domain.WriteResponse(w, http.StatusOK, resp); err != nil {
			slog.Error("failed to write response", "error", err)
			domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)

			return
		}
	}
}
