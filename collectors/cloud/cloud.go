package cloud

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"confirmate.io/collectors/cloud/internal/collector"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// Config is the public collector configuration.
type Config = collector.Config

// DefaultConfig returns sensible defaults for non-secret settings.
func DefaultConfig() Config {
	return collector.DefaultConfig()
}

// LoadConfigFromEnv reads collector configuration from environment variables.
func LoadConfigFromEnv() (Config, error) {
	return collector.LoadConfigFromEnv()
}

// Start starts the Azure collector and blocks until ctx is canceled.
func Start(ctx context.Context, cfg Config, logger *slog.Logger) error {
	var (
		err     error
		cred    *azidentity.DefaultAzureCredential
		fetcher *collector.AzureVMFetcher
	)

	if err = cfg.Validate(); err != nil {
		return err
	}

	if logger == nil {
		logger = slog.Default()
	}

	if cfg.AzureSubscriptionID == "" {
		cfg.AzureSubscriptionID, err = collector.ResolveSubscriptionIDFromAzureCLI(ctx)
		if err != nil {
			return fmt.Errorf("failed to resolve azure subscription id automatically: %w", err)
		}
		logger.Info("resolved azure subscription id from az cli context")
	}

	cred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize azure credential: %w", err)
	}

	fetcher, err = collector.NewAzureVMFetcher(cfg.AzureSubscriptionID, cred, cfg.AzureResourceGroup, nil)
	if err != nil {
		return fmt.Errorf("failed to create azure fetcher: %w", err)
	}

	httpClient := &http.Client{Timeout: cfg.HTTPTimeout}
	sink := collector.NewEvidenceStoreSink(httpClient, cfg.EvidenceStoreAddress)
	svc := collector.NewService(cfg, fetcher, sink, logger)

	return svc.Start(ctx)
}
