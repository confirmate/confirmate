package collector

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"confirmate.io/core/api/evidence"
)

type fakeFetcher struct {
	vms []AzureVirtualMachine
	err error
}

func (f fakeFetcher) ListVirtualMachines(_ context.Context) ([]AzureVirtualMachine, error) {
	return f.vms, f.err
}

type fakeSink struct {
	stored []*evidence.Evidence
	errAt  map[string]error
}

func (f *fakeSink) StoreEvidence(_ context.Context, ev *evidence.Evidence) error {
	if err, ok := f.errAt[ev.GetResource().GetVirtualMachine().GetId()]; ok {
		return err
	}
	f.stored = append(f.stored, ev)
	return nil
}

func TestService_RunCycle(t *testing.T) {
	cfg := Config{
		TargetOfEvaluationID: "11111111-1111-1111-1111-111111111111",
		ToolID:               "cloud-collector-azure",
	}

	sink := &fakeSink{}
	svc := NewService(cfg, fakeFetcher{vms: []AzureVirtualMachine{{ID: "vm-1", Name: "vm-1"}, {ID: "vm-2", Name: "vm-2"}}}, sink, slog.New(slog.NewTextHandler(io.Discard, nil)))
	svc.nowFn = func() time.Time { return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC) }
	svc.uuidFn = func() string { return "22222222-2222-2222-2222-222222222222" }

	summary, err := svc.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle() error = %v", err)
	}

	if summary.FetchedVMs != 2 || summary.SentEvidences != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(sink.stored) != 2 {
		t.Fatalf("expected 2 stored evidences, got %d", len(sink.stored))
	}
}

func TestService_RunCycle_PartialSubmissionFailure(t *testing.T) {
	cfg := Config{
		TargetOfEvaluationID: "11111111-1111-1111-1111-111111111111",
		ToolID:               "cloud-collector-azure",
	}

	sink := &fakeSink{errAt: map[string]error{"vm-2": errors.New("sink failure")}}
	svc := NewService(cfg, fakeFetcher{vms: []AzureVirtualMachine{{ID: "vm-1", Name: "vm-1"}, {ID: "vm-2", Name: "vm-2"}}}, sink, slog.New(slog.NewTextHandler(io.Discard, nil)))
	svc.nowFn = func() time.Time { return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC) }
	svc.uuidFn = func() string { return "22222222-2222-2222-2222-222222222222" }

	summary, err := svc.RunCycle(context.Background())
	if err == nil {
		t.Fatal("expected RunCycle() to return error")
	}

	if summary.FetchedVMs != 2 {
		t.Fatalf("unexpected fetched count: %+v", summary)
	}
	if summary.SentEvidences != 1 {
		t.Fatalf("unexpected sent count: %+v", summary)
	}
	if summary.FailedSubmissions != 1 {
		t.Fatalf("unexpected failed submissions count: %+v", summary)
	}
}

func TestService_RunCycle_FetchFailure(t *testing.T) {
	cfg := Config{
		TargetOfEvaluationID: "11111111-1111-1111-1111-111111111111",
		ToolID:               "cloud-collector-azure",
	}

	svc := NewService(cfg, fakeFetcher{err: errors.New("fetch error")}, &fakeSink{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	summary, err := svc.RunCycle(context.Background())
	if err == nil {
		t.Fatal("expected RunCycle() to return error")
	}
	if summary.FetchedVMs != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
