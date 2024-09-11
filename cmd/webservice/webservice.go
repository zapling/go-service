package webservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var addr = ":3000"

// Run starts the http server while listening for a context cancelation on the provided
// context. If the context is cancelled it will try and gracefully shutdown the http server.
func Run(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	router, err := NewRouter(ctx)
	if err != nil {
		return fmt.Errorf("failed to get router: %w", err)
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msgf("Starting http server on %s", addr)
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Error while trying to serve")
		}
	}()

	// Graceful shutdown
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		err := httpServer.Shutdown(ctxWithTimeout)
		if err != nil {
			log.Err(err).Msg("Error while trying to gracefully shutdown")
		}
	}()

	wg.Wait()

	return nil
}
