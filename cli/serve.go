package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/server"
	"github.com/spf13/cobra"
)

var (
	debugMode bool
	addr      string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the AI Router server",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug mode: log full request/response lifecycle for every API call")
	serveCmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8787", "Address to listen on")
}

func runServe(_ *cobra.Command, _ []string) error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	providers := verifiedProviders(ctx)
	useCases := useCases(providers)
	for _, uc := range useCases {
		uc.Start(ctx)
	}

	cfg := server.Config{Addr: addr, Providers: providers, UseCases: useCases}
	if debugMode {
		slog.Info("Debug mode enabled — full request/response lifecycle will be logged")
		cfg.Debug = os.Stdout
	}
	srv := server.New(cfg)

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	return srv.Start()
}

// verifiedProviders builds every configured provider and keeps those that
// pass verification, sorted by name.
func verifiedProviders(ctx context.Context) []provider.Provider {
	var candidates []provider.Provider
	for name, build := range cloudProviders {
		// Env var format: AI_ROUTER_<PROVIDER>_API_KEY (e.g. AI_ROUTER_OPENROUTER_API_KEY)
		if key := os.Getenv(fmt.Sprintf("AI_ROUTER_%s_API_KEY", strings.ToUpper(name))); key != "" {
			candidates = append(candidates, build(key))
		}
	}
	for _, build := range localProviders {
		candidates = append(candidates, build())
	}

	var (
		mu       sync.Mutex
		verified []provider.Provider
		wg       sync.WaitGroup
	)
	for _, p := range candidates {
		wg.Go(func() {
			// A local runner that is not running is normal, so it gets a short
			// timeout and no warning.
			timeout, logFailure := 10*time.Second, slog.Warn
			if p.Type() == provider.Local {
				timeout, logFailure = 2*time.Second, slog.Debug
			}
			verifyCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			if err := p.Verify(verifyCtx); err != nil {
				logFailure("Provider verification failed, skipping", "provider", p.Name(), "error", err)
				return
			}
			slog.Info("Registered provider", "provider", p.Name(), "type", p.Type())
			mu.Lock()
			verified = append(verified, p)
			mu.Unlock()
		})
	}
	wg.Wait()

	sort.Slice(verified, func(i, j int) bool { return verified[i].Name() < verified[j].Name() })
	return verified
}
