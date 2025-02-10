package routes

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(ctx context.Context, logger *slog.Logger, router chi.Router) (err error) {

	if err := errors.Join(
		setupIndexRoute(router),
		// Add more routes here
	); err != nil {
		return fmt.Errorf("error setting up routes: %w", err)
	}

	return nil
}
