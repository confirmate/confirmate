package collector

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	envEvidenceStoreAddress = "CLOUD_COLLECTOR_EVIDENCE_STORE_ADDR"
	envTargetOfEvaluationID = "CLOUD_COLLECTOR_TARGET_OF_EVALUATION_ID"
	envToolID               = "CLOUD_COLLECTOR_TOOL_ID"
	envAzureSubscriptionID  = "CLOUD_COLLECTOR_AZURE_SUBSCRIPTION_ID"
	envAzureResourceGroup   = "CLOUD_COLLECTOR_AZURE_RESOURCE_GROUP"
	envInterval             = "CLOUD_COLLECTOR_INTERVAL"
	envCycleTimeout         = "CLOUD_COLLECTOR_CYCLE_TIMEOUT"
	envHTTPTimeout          = "CLOUD_COLLECTOR_HTTP_TIMEOUT"

	envLegacyTargetOfEvaluationID = "TARGET_OF_EVALUATION_ID"
	envLegacyAzureSubscriptionID  = "AZURE_SUBSCRIPTION_ID"
	envLegacyAzureResourceGroup   = "AZURE_RESOURCE_GROUP"

	defaultEvidenceStoreAddress = "http://localhost:8080"
	defaultToolID               = "cloud-collector-azure"
	defaultInterval             = 5 * time.Minute
	defaultCycleTimeout         = 60 * time.Second
	defaultHTTPTimeout          = 15 * time.Second
)

// Config holds collector runtime and integration settings.
type Config struct {
	EvidenceStoreAddress string
	TargetOfEvaluationID string
	ToolID               string
	AzureSubscriptionID  string
	AzureResourceGroup   string
	Interval             time.Duration
	CycleTimeout         time.Duration
	HTTPTimeout          time.Duration
}

// DefaultConfig returns sensible defaults for non-secret collector settings.
func DefaultConfig() Config {
	return Config{
		EvidenceStoreAddress: defaultEvidenceStoreAddress,
		ToolID:               defaultToolID,
		Interval:             defaultInterval,
		CycleTimeout:         defaultCycleTimeout,
		HTTPTimeout:          defaultHTTPTimeout,
	}
}

// LoadConfigFromEnv reads and validates collector configuration from environment variables.
func LoadConfigFromEnv() (cfg Config, err error) {
	cfg = DefaultConfig()

	cfg.EvidenceStoreAddress = firstNonEmpty(
		trimEnv(os.Getenv(envEvidenceStoreAddress)),
		cfg.EvidenceStoreAddress,
	)
	cfg.TargetOfEvaluationID = firstNonEmpty(
		trimEnv(os.Getenv(envTargetOfEvaluationID)),
		trimEnv(os.Getenv(envLegacyTargetOfEvaluationID)),
	)
	cfg.ToolID = firstNonEmpty(
		trimEnv(os.Getenv(envToolID)),
		cfg.ToolID,
	)
	cfg.AzureSubscriptionID = firstNonEmpty(
		trimEnv(os.Getenv(envAzureSubscriptionID)),
		trimEnv(os.Getenv(envLegacyAzureSubscriptionID)),
	)
	cfg.AzureResourceGroup = firstNonEmpty(
		trimEnv(os.Getenv(envAzureResourceGroup)),
		trimEnv(os.Getenv(envLegacyAzureResourceGroup)),
	)

	if raw := trimEnv(os.Getenv(envInterval)); raw != "" {
		cfg.Interval, err = time.ParseDuration(raw)
		if err != nil {
			return cfg, fmt.Errorf("invalid %s: %w", envInterval, err)
		}
	}

	if raw := trimEnv(os.Getenv(envCycleTimeout)); raw != "" {
		cfg.CycleTimeout, err = time.ParseDuration(raw)
		if err != nil {
			return cfg, fmt.Errorf("invalid %s: %w", envCycleTimeout, err)
		}
	}

	if raw := trimEnv(os.Getenv(envHTTPTimeout)); raw != "" {
		cfg.HTTPTimeout, err = time.ParseDuration(raw)
		if err != nil {
			return cfg, fmt.Errorf("invalid %s: %w", envHTTPTimeout, err)
		}
	}

	if err = cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Validate validates configuration values.
func (cfg Config) Validate() error {
	if cfg.EvidenceStoreAddress == "" {
		return fmt.Errorf("%s is required", envEvidenceStoreAddress)
	}
	if cfg.TargetOfEvaluationID == "" {
		return fmt.Errorf("%s is required", envTargetOfEvaluationID)
	}
	if _, err := uuid.Parse(cfg.TargetOfEvaluationID); err != nil {
		return fmt.Errorf("%s must be a UUID: %w", envTargetOfEvaluationID, err)
	}
	if cfg.ToolID == "" {
		return fmt.Errorf("%s must not be empty", envToolID)
	}
	if cfg.Interval <= 0 {
		return fmt.Errorf("%s must be > 0", envInterval)
	}
	if cfg.CycleTimeout <= 0 {
		return fmt.Errorf("%s must be > 0", envCycleTimeout)
	}
	if cfg.HTTPTimeout <= 0 {
		return fmt.Errorf("%s must be > 0", envHTTPTimeout)
	}
	if cfg.CycleTimeout > cfg.Interval {
		return fmt.Errorf("%s must be <= %s", envCycleTimeout, envInterval)
	}

	return nil
}

func trimEnv(s string) string {
	return strings.TrimSpace(s)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
