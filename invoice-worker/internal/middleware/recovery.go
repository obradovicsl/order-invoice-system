package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	httputils "invoice-worker/internal/utils/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RecoveryMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := chimiddleware.GetReqID(request.Context())

					logger.Error("panic recovered",
						"request_id", requestID,
						"error", err,
						"stack", string(debug.Stack()),
					)

					_ = httputils.WriteErrorResponse(writer, http.StatusInternalServerError, "Internal Server Error", httputils.ErrorCodeInternalServerError)
				}
			}()

			next.ServeHTTP(writer, request)
		})
	}
}
