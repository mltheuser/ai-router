package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the AI Router server",
	RunE:  runServe,
}

func runServe(cmd *cobra.Command, args []string) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Set up registry
	reg := router.DefaultRegistry()

	// Collect providers in parallel
	var providers []provider.Provider
	var providersMu sync.Mutex

	// Register cloud and local providers in parallel
	var wg sync.WaitGroup

	// Cloud providers
	for _, name := range reg.CloudNames() {
		// Env var format: AI_ROUTER_<PROVIDER>_API_KEY (e.g. AI_ROUTER_OPENROUTER_API_KEY)
		envKey := fmt.Sprintf("AI_ROUTER_%s_API_KEY", strings.ToUpper(name))
		if key := os.Getenv(envKey); key != "" {
			wg.Add(1)
			go func(name, key string) {
				defer wg.Done()

				p, err := reg.NewCloud(name, key)
				if err != nil {
					logger.Warn("Failed to create cloud provider", "provider", name, "error", err)
					return
				}
				
				// Verify before registering
				verifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				err = p.Verify(verifyCtx)
				cancel()

				if err != nil {
					logger.Warn("Cloud provider verification failed, skipping", "provider", name, "error", err)
					return
				}

				logger.Info("Registered cloud provider", "provider", name)
				providersMu.Lock()
				providers = append(providers, p)
				providersMu.Unlock()
			}(name, key)
		}
	}

	// Local runners
	for _, name := range reg.LocalNames() {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			// Create with default endpoint (provider handles default)
			p, err := reg.NewLocal(name)
			if err != nil {
				logger.Warn("Failed to create local runner", "runner", name, "error", err)
				return
			}

			// Verify with short timeout
			verifyCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := p.Verify(verifyCtx); err != nil {
				// Just ignore failure as per instructions
				logger.Debug("Local runner not available", "runner", name, "error", err)
				return
			}

			logger.Info("Registered local runner", "runner", name)
			providersMu.Lock()
			providers = append(providers, p)
			providersMu.Unlock()
		}(name)
	}
	wg.Wait()

	// Set up the model catalog with all verified providers
	catalog := router.NewModelCatalog(logger, providers)

	// Initialize catalog (parallel poll of all providers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("Polling providers for available models...")
	// We'll trust Initialize to handle verification and skipping unavailable providers
	if err := catalog.Initialize(ctx); err != nil { 
		logger.Warn("Catalog initialization issue", "error", err)
	}

	// Start background refresh
	catalog.StartBackgroundRefresh(ctx)

	// Create router and server
	r := router.NewRouter(catalog)
	srv := server.New("127.0.0.1", 8787, r, catalog, logger)

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Received shutdown signal")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second) 
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	return srv.Start()
}
