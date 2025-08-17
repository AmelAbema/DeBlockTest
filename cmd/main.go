package main

import (
	"DeBlockTest/cmd/app"
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tel-io/tel/v2"
)

func main() {
	ccx, cancel := context.WithCancel(context.Background())

	cfg := tel.GetConfigFromEnv()
	t, cc := tel.New(ccx, cfg)
	ctx := tel.WithContext(ccx, t)

	done := make(chan struct{})

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, os.Interrupt)
		<-ch
		t.Info("shutdown signal received, starting graceful shutdown")

		cc()
		cancel()

		select {
		case <-done:
			t.Info("graceful shutdown completed")
		case <-time.After(3 * time.Second):
			t.Warn("shutdown timeout exceeded, forcing exit")
		}
		os.Exit(0)
	}()

	server := app.New()

	t.Info("starting DeBlock monitoring service")

	if err := server.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal("service failed:", tel.Error(err))
	}

	close(done)
	t.Info("service stopped")
}
