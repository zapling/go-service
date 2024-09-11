package webservice

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/zapling/go-service/cmd/webservice/handler"
)

// attachMiddleware wraps the provided middleware functions around the provided root handler.
func attachMiddleware(root http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	var handler http.Handler = root

	// Reverse middleware order so we wrap them in the order they where defined.
	slices.Reverse(middlewares)

	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

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
			if logger == nil {
				next.ServeHTTP(w, r)
				return
			}

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

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// logAccess logs every incoming request and outgoing response.
//
// Any unrecovered panics will be logged as error with the corresponding stack trace.
//
// The log level for the response log is determined by the status code of the response.
// Default: Info
// 100 - 199: Debug
// 200 - 399: Info
// 400 - 499: Warn
// 500 - 599: Error
// No status code: Error
func logAccess(logger *zerolog.Logger) func(http.Handler) http.Handler {

	getLogEvent := func(statusCode int) *zerolog.Event {
		switch statusCode := statusCode; {
		case statusCode == 0:
			return logger.Error()
		case statusCode >= 100 && statusCode <= 199:
			return logger.Debug()
		case statusCode >= 200 && statusCode <= 399:
			return logger.Info()
		case statusCode >= 400 && statusCode <= 499:
			return logger.Warn()
		case statusCode >= 500 && statusCode <= 599:
			return logger.Error()
		default:
			return logger.Info()
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := &loggingResponseWriter{ResponseWriter: w}

			logger.Info().
				Fields(map[string]any{
					"url":        r.URL.Path,
					"proto":      r.Proto,
					"method":     r.Method,
					"remote-ip":  r.RemoteAddr,
					"user-agent": r.Header.Get("User-Agent"),
					"trace-id":   r.Header.Get(requestTraceHeaderKey),
				}).Msg("Incoming request")

			defer func() {
				end := time.Now()

				if rec := recover(); rec != nil {
					logger.Error().
						Interface("recover-info", rec).
						Bytes("stack-trace", debug.Stack()).
						Str("trace-id", r.Header.Get(requestTraceHeaderKey)).
						Msg("Panic while handling incoming request")
				}

				log := getLogEvent(lrw.statusCode)
				log.Fields(map[string]any{
					"status":     lrw.statusCode,
					"trace-id":   r.Header.Get(requestTraceHeaderKey),
					"latency-ms": float64(end.Sub(start).Nanoseconds()) / 1000000.0,
				}).Msg("Outgoing response")
			}()

			next.ServeHTTP(lrw, r)
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
