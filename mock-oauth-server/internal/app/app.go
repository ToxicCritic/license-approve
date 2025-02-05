package app

import (
	"crypto/tls"
	"fmt"
	"log"
	"mock-oauth-server/config"
	"mock-oauth-server/internal/api/v1/httpapi/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	cfg    *config.Config
	server *http.Server
}

func New(cfg *config.Config) (*App, error) {
	r := router.New()

	if _, err := os.Stat(cfg.CertFile); err != nil {
		return nil, fmt.Errorf("cert file not found: %v", err)
	}
	if _, err := os.Stat(cfg.KeyFile); err != nil {
		return nil, fmt.Errorf("key file not found: %v", err)
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	app := &App{
		cfg:    cfg,
		server: srv,
	}
	return app, nil
}

func (a *App) Run() error {
	// Ловим сигналы для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("[App] Shutting down mock-oauth-server gracefully...")
		if err := a.server.Close(); err != nil {
			log.Printf("[App] Server Close error: %v\n", err)
		}
	}()

	log.Printf("[App] Starting mock OAuth2.0 server on %s...\n", a.cfg.Addr)

	err := a.server.ListenAndServeTLS(a.cfg.CertFile, a.cfg.KeyFile)
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("ListenAndServeTLS error: %w", err)
	}
	time.Sleep(1 * time.Second)
	log.Println("[App] Server stopped.")
	return nil
}
