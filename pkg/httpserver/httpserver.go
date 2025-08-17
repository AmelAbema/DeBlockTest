package httpserver

import (
	"DeBlockTest/internal/config"
	"DeBlockTest/pkg/addresses"
	"DeBlockTest/pkg/processing"
	"DeBlockTest/pkg/transport"
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
)

type HTTPServer struct {
	server *http.Server
	config *config.HTTPConfig
}

func NewHTTPServer(
	cfg *config.HTTPConfig,
	addresses *addresses.AddressModule,
	processing *processing.ProcessingModule,
) *HTTPServer {
	mux := http.NewServeMux()

	monitoringAPI := transport.NewMonitoringAPI(addresses, processing)
	monitoringAPI.RegisterHandlers(mux)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &HTTPServer{
		server: server,
		config: cfg,
	}
}

func (h *HTTPServer) Start(ctx context.Context) error {
	tel.Global().Info("starting HTTP server",
		tel.String("address", h.config.Address))

	go func() {
		<-ctx.Done()
		tel.Global().Info("shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.server.Shutdown(shutdownCtx); err != nil {
			tel.Global().Error("HTTP server shutdown error", tel.Error(err))
		} else {
			tel.Global().Info("HTTP server shutdown completed")
		}
	}()

	if err := h.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "HTTP server failed")
	}

	return nil
}

func (h *HTTPServer) Stop(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}
