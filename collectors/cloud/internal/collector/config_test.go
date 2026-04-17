package collector

import (
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv(envEvidenceStoreAddress, "http://localhost:9091")
	t.Setenv(envTargetOfEvaluationID, "11111111-1111-1111-1111-111111111111")
	t.Setenv(envToolID, "cloud-collector")
	t.Setenv(envAzureSubscriptionID, "sub-123")
	t.Setenv(envAzureResourceGroup, "rg-1")
	t.Setenv(envInterval, "2m")
	t.Setenv(envCycleTimeout, "30s")
	t.Setenv(envHTTPTimeout, "10s")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}

	if cfg.EvidenceStoreAddress != "http://localhost:9091" {
		t.Fatalf("unexpected evidence store address: %q", cfg.EvidenceStoreAddress)
	}
	if cfg.TargetOfEvaluationID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected target of evaluation id: %q", cfg.TargetOfEvaluationID)
	}
	if cfg.ToolID != "cloud-collector" {
		t.Fatalf("unexpected tool id: %q", cfg.ToolID)
	}
	if cfg.AzureSubscriptionID != "sub-123" {
		t.Fatalf("unexpected subscription id: %q", cfg.AzureSubscriptionID)
	}
	if cfg.AzureResourceGroup != "rg-1" {
		t.Fatalf("unexpected resource group: %q", cfg.AzureResourceGroup)
	}
	if cfg.Interval != 2*time.Minute {
		t.Fatalf("unexpected interval: %s", cfg.Interval)
	}
	if cfg.CycleTimeout != 30*time.Second {
		t.Fatalf("unexpected cycle timeout: %s", cfg.CycleTimeout)
	}
	if cfg.HTTPTimeout != 10*time.Second {
		t.Fatalf("unexpected http timeout: %s", cfg.HTTPTimeout)
	}
}

func TestLoadConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv(envTargetOfEvaluationID, "11111111-1111-1111-1111-111111111111")
	t.Setenv(envAzureSubscriptionID, "sub-123")
	t.Setenv(envAzureResourceGroup, "rg-1")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}

	if cfg.EvidenceStoreAddress != defaultEvidenceStoreAddress {
		t.Fatalf("unexpected default evidence store address: %q", cfg.EvidenceStoreAddress)
	}
	if cfg.ToolID != defaultToolID {
		t.Fatalf("unexpected default tool id: %q", cfg.ToolID)
	}
	if cfg.Interval != defaultInterval {
		t.Fatalf("unexpected default interval: %s", cfg.Interval)
	}
	if cfg.CycleTimeout != defaultCycleTimeout {
		t.Fatalf("unexpected default cycle timeout: %s", cfg.CycleTimeout)
	}
	if cfg.HTTPTimeout != defaultHTTPTimeout {
		t.Fatalf("unexpected default http timeout: %s", cfg.HTTPTimeout)
	}
}

func TestLoadConfigFromEnv_ValidationError(t *testing.T) {
	t.Setenv(envTargetOfEvaluationID, "not-a-uuid")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoadConfigFromEnv_SubscriptionAndResourceGroupOptional(t *testing.T) {
	t.Setenv(envTargetOfEvaluationID, "11111111-1111-1111-1111-111111111111")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}

	if cfg.AzureSubscriptionID != "" {
		t.Fatalf("expected empty subscription id, got %q", cfg.AzureSubscriptionID)
	}
	if cfg.AzureResourceGroup != "" {
		t.Fatalf("expected empty resource group, got %q", cfg.AzureResourceGroup)
	}
}

func TestLoadConfigFromEnv_LegacyEnvFallbacks(t *testing.T) {
	t.Setenv(envLegacyTargetOfEvaluationID, "11111111-1111-1111-1111-111111111111")
	t.Setenv(envLegacyAzureSubscriptionID, "sub-legacy")
	t.Setenv(envLegacyAzureResourceGroup, "rg-legacy")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}

	if cfg.TargetOfEvaluationID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected target of evaluation id: %q", cfg.TargetOfEvaluationID)
	}
	if cfg.AzureSubscriptionID != "sub-legacy" {
		t.Fatalf("unexpected subscription id: %q", cfg.AzureSubscriptionID)
	}
	if cfg.AzureResourceGroup != "rg-legacy" {
		t.Fatalf("unexpected resource group: %q", cfg.AzureResourceGroup)
	}
}
