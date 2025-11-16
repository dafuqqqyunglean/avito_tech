package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/dafuqqqyunglean/avito_tech/domain"
)

func RecoveryMiddleware(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic occurred: %v\n%s", err, debug.Stack())
				domain.NewErrorResponse(ctx, w, domain.ErrInternal, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
