package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ainghazal/tunnel-telemetry/internal/collector"
	"github.com/ainghazal/tunnel-telemetry/internal/server"
)

func main() {
	// TODO(ain): pass config (viper)
	cfg := &server.Config{
		DebugGeolocation: true,
	}

	e := server.NewEchoServer(cfg)

	collector := &collector.FileSystemCollector{}
	h := server.NewHandler(collector)

	e.GET("/", server.HandleRootDecoy)
	e.POST("/report", h.CreateReport)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("shutting down the server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
