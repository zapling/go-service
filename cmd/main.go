package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/zapling/go-service/cmd/webservice"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)

	if err := run(ctx); err != nil {
		logger.Err(err).Msg("Error while running, exiting")
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var component string
	if len(os.Args) > 1 {
		component = os.Args[1]
	}

	ctx = zerolog.Ctx(ctx).
		With().
		Str("component", component).
		Logger().
		WithContext(ctx)

	switch component {
	case "webservice":
		return webservice.Run(ctx)
	default:
		return fmt.Errorf("unknown component '%s', failed to start", component)
	}
}
