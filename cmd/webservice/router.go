package webservice

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/zapling/go-service/cmd/webservice/handler"
	"github.com/zapling/go-service/internal/business"
)

func newRouter(ctx context.Context) (http.Handler, error) {
	log := zerolog.Ctx(ctx)

	r := http.NewServeMux()

	bc := business.New()
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

func attachRoutes(r *http.ServeMux, bc *business.Client) {
	r.Handle("GET /foo", handler.Foo(bc))
}
