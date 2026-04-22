package collector

import (
	"context"

	"confirmate.io/core/api/evidence"
)

// AzureVirtualMachine is the minimal provider DTO for the MVP collector.
type AzureVirtualMachine struct {
	ID                  string
	Name                string
	Region              string
	Tags                map[string]string
	BootLoggingEnabled  *bool
	BootLoggingStoreURI string
}

// VMsFetcher retrieves Azure VM inventory.
type VMsFetcher interface {
	ListVirtualMachines(ctx context.Context) ([]AzureVirtualMachine, error)
}

// EvidenceSink stores evidence records in the evidence store service.
type EvidenceSink interface {
	StoreEvidence(ctx context.Context, evidence *evidence.Evidence) error
}
