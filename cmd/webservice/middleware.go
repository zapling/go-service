package webservice

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/zapling/go-service/cmd/webservice/handler"
)

const requestTraceHeaderKey = "x-trace-id"

// setRequestTraceIdHeader generates a uuidv4 trace-id and updates the setRequestTraceIdHeader
// header if there is not one already set.
func setRequestTraceIdHeader() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(requestTraceHeaderKey) == "" {
				r.Header.Set(requestTraceHeaderKey, uuid.New().String())
			}
			next.ServeHTTP(w, r)
		})
	}
}

// setRequestContextLogger creates a new logger and attaches it to the request context,
// making the logger available for next handlers. If logWithTraceId is enabled, a trace-id key
// value pair will be logged with every message, making it easier to correlate related logs.
func setRequestContextLogger(logger *zerolog.Logger, logWithTraceId bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := *logger
			if logWithTraceId {
				traceId := r.Header.Get(requestTraceHeaderKey)
				if traceId != "" {
					log = log.With().Str("trace-id", traceId).Logger()
				}
			}
			ctx := log.WithContext(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

const localhostOrigin = "http://localhost"

// setResponseCORSHeaders sets the necessary headers to allow for cross origin requests.
func setResponseCORSHeaders(
	allowedOrigins []string,
	allowedHeaders []string,
	allowedMethods []string,
	allowCredentials bool,
) func(http.Handler) http.Handler {

	getAllowedOrigin := func(r *http.Request) string {
		origin := r.Header.Get("origin")
		for _, allowedOrigin := range allowedOrigins {
			// If we allow localhost, we allow it from any port
			if allowedOrigin == localhostOrigin && strings.Contains(origin, allowedOrigin) {
				u, _ := url.Parse(origin)
				port := u.Port()
				if port != "" {
					return allowedOrigin + ":" + port
				}
				return allowedOrigin
			}

			if allowedOrigin == origin {
				return allowedOrigin
			}
		}
		return "" // not allowed
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// We do not care about CORS headers if there is no origin header.
			// If the client enforces CORS it should also send an origin header, right?
			origin := r.Header.Get("origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			allowedOrigin := getAllowedOrigin(r)
			if allowedOrigin == "" {
				zerolog.Ctx(r.Context()).Warn().Msg("Request origin does not match any of the allowed origins")
				handler.WriteJson(w, r, http.StatusForbidden, fmt.Errorf("origin '%s' is not allowed", origin))
				return
			}

			w.Header().Set("access-control-allow-origin", allowedOrigin)
			w.Header().Set("access-control-allow-headers", strings.Join(allowedHeaders, ", "))
			w.Header().Set("access-control-allow-methods", strings.Join(allowedMethods, ", "))

			if allowCredentials {
				w.Header().Set("access-control-allow-credentials", "true")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
