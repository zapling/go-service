package webservice

import (
	"context"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/zapling/go-service/cmd/webservice/handler"
	"github.com/zapling/go-service/internal/business"
)

// newRouter creates a new http mux with any routes that it needs.
// Any dependencies you might need inside your handlers should be instantiated here
// e.g database connection, queue connection etc
// Use environment variables or mounted secrets (k8s) to access the credentials you
// might need.
func newRouter(ctx context.Context) (http.Handler, error) {
	log := zerolog.Ctx(ctx)

	db, err := getDatabaseConn(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	bc := business.New(db)

	r := http.NewServeMux()
	attachRoutes(r, bc)

	return attachMiddleware(
		r,
		setRequestTraceIdHeader(),
		setRequestContextLogger(log, true),
		setResponseCORSHeaders(
			[]string{localhostOrigin, "https://mywebsitedomain.com"},
			[]string{"content-type", "authorization"},
			[]string{"HEAD", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"},
			false,
		),
	), nil
}

// attachRoutes attaches routes to the provided http mux.
func attachRoutes(r *http.ServeMux, bc *business.Client) {
	r.Handle("GET /foo", handler.Foo(bc))
}
