package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"confirmate.io/core/api/evidence"

	"github.com/google/uuid"
)

// Summary contains one collection-cycle result.
type Summary struct {
	FetchedVMs        int
	SentEvidences     int
	FailedTransforms  int
	FailedSubmissions int
}

// Service orchestrates fetching VMs, mapping evidence, and submitting it.
type Service struct {
	cfg     Config
	fetcher VMsFetcher
	sink    EvidenceSink
	logger  *slog.Logger
	nowFn   func() time.Time
	uuidFn  func() string
}

// NewService creates a new collector service.
func NewService(cfg Config, fetcher VMsFetcher, sink EvidenceSink, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		cfg:     cfg,
		fetcher: fetcher,
		sink:    sink,
		logger:  logger,
		nowFn:   time.Now,
		uuidFn:  uuid.NewString,
	}
}

// RunCycle executes one collection cycle.
func (s *Service) RunCycle(ctx context.Context) (summary Summary, err error) {
	var (
		vms  []AzureVirtualMachine
		errs []error
		evid *evidence.Evidence
	)

	vms, err = s.fetcher.ListVirtualMachines(ctx)
	if err != nil {
		return summary, fmt.Errorf("fetching azure VMs: %w", err)
	}

	summary.FetchedVMs = len(vms)

	for _, vm := range vms {
		evid, err = MapVMToEvidence(vm, s.cfg, s.uuidFn(), s.nowFn())
		if err != nil {
			summary.FailedTransforms++
			errs = append(errs, fmt.Errorf("mapping VM %q to evidence: %w", vm.ID, err))
			continue
		}

		err = s.sink.StoreEvidence(ctx, evid)
		if err != nil {
			summary.FailedSubmissions++
			errs = append(errs, fmt.Errorf("submitting VM %q evidence: %w", vm.ID, err))
			continue
		}

		summary.SentEvidences++
	}

	if len(errs) > 0 {
		return summary, errors.Join(errs...)
	}

	return summary, nil
}

// Start runs collection immediately and then in configured intervals until ctx is canceled.
func (s *Service) Start(ctx context.Context) error {
	run := func() {
		cycleCtx, cancel := context.WithTimeout(ctx, s.cfg.CycleTimeout)
		defer cancel()

		summary, err := s.RunCycle(cycleCtx)
		if err != nil {
			s.logger.Error("collection cycle finished with errors",
				slog.Int("fetched_vms", summary.FetchedVMs),
				slog.Int("sent_evidences", summary.SentEvidences),
				slog.Int("failed_transforms", summary.FailedTransforms),
				slog.Int("failed_submissions", summary.FailedSubmissions),
				slog.String("error", err.Error()),
			)
			return
		}

		s.logger.Info("collection cycle finished",
			slog.Int("fetched_vms", summary.FetchedVMs),
			slog.Int("sent_evidences", summary.SentEvidences),
		)
	}

	run()

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			run()
		}
	}
}
