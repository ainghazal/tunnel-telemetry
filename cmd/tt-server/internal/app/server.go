package app

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	"github.com/ainghazal/tunnel-telemetry/internal/collector"
	"github.com/ainghazal/tunnel-telemetry/internal/config"
	"github.com/ainghazal/tunnel-telemetry/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var commitInfo = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:12]
			}
		}
	}
	return "unknown"
}()

func handleVersionInfo(c echo.Context) error {
	return c.String(http.StatusOK, commitInfo)
}

func startEchoServer(cfg *config.Config) {
	e := server.NewEchoServer(cfg)
	if cfg.Debug {
		e.Logger.SetLevel(log.DEBUG)
	}

	collector := collector.NewFileSystemCollector(cfg)
	h := server.NewHandler(collector, collector)

	e.GET("/", server.HandleRootDecoy)
	e.POST("/report", h.CreateReport)
	e.GET("/version", handleVersionInfo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if cfg.AutoTLS {
		// Start server
		go startAutoTLSServer(e, cfg)
	} else {
		go func() {
			if err := e.Start(cfg.ListenAddr); err != nil && err != http.ErrServerClosed {
				e.Logger.Fatalf("shutting down the server: %v", err)
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func startAutoTLSServer(e *echo.Echo, cfg *config.Config) {
	autoTLSManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		// Cache certificates to avoid issues with rate limits (https://letsencrypt.org/docs/rate-limits)
		Cache:      autocert.DirCache(cfg.AutoTLSCacheDir),
		HostPolicy: autocert.HostWhitelist(cfg.Hostname),
	}
	s := http.Server{
		Addr:    cfg.ListenAddr,
		Handler: e, // set Echo as handler
		TLSConfig: &tls.Config{
			GetCertificate: autoTLSManager.GetCertificate,
			NextProtos:     []string{acme.ALPNProto},
		},
		ReadTimeout: 30 * time.Second, // use custom timeouts
	}
	e.Logger.Info("Starting autotls server (can take a few secs)")
	if err := s.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
