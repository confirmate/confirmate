package collector

import (
	"testing"
	"time"
)

func TestMapVMToEvidence(t *testing.T) {
	now := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	bootLoggingEnabled := true
	cfg := Config{
		TargetOfEvaluationID: "11111111-1111-1111-1111-111111111111",
		ToolID:               "cloud-collector-azure",
	}

	ev, err := MapVMToEvidence(AzureVirtualMachine{
		ID:                  "/subscriptions/sub-123/resourceGroups/rg-1/providers/Microsoft.Compute/virtualMachines/vm-1",
		Name:                "vm-1",
		Tags:                map[string]string{"env": "dev"},
		BootLoggingEnabled:  &bootLoggingEnabled,
		BootLoggingStoreURI: "https://example.blob.core.windows.net/bootdiagnostics",
	}, cfg, "22222222-2222-2222-2222-222222222222", now)
	if err != nil {
		t.Fatalf("MapVMToEvidence() error = %v", err)
	}

	if ev.GetId() == "" {
		t.Fatal("expected evidence ID to be set")
	}
	if ev.GetTimestamp() == nil || !ev.GetTimestamp().AsTime().Equal(now) {
		t.Fatalf("unexpected timestamp: %v", ev.GetTimestamp())
	}
	if ev.GetTargetOfEvaluationId() != cfg.TargetOfEvaluationID {
		t.Fatalf("unexpected target of evaluation id: %q", ev.GetTargetOfEvaluationId())
	}
	if ev.GetToolId() != cfg.ToolID {
		t.Fatalf("unexpected tool id: %q", ev.GetToolId())
	}
	if ev.GetResource() == nil {
		t.Fatal("expected resource to be set")
	}
	vm := ev.GetResource().GetVirtualMachine()
	if vm == nil {
		t.Fatal("expected virtual_machine resource to be set")
	}
	if vm.GetId() == "" || vm.GetName() == "" {
		t.Fatalf("expected VM id/name to be set, got id=%q name=%q", vm.GetId(), vm.GetName())
	}
	if vm.GetLabels()["env"] != "dev" {
		t.Fatalf("unexpected labels: %v", vm.GetLabels())
	}
	if vm.GetBootLogging() == nil {
		t.Fatal("expected boot logging to be set")
	}
	if !vm.GetBootLogging().GetEnabled() {
		t.Fatal("expected boot logging enabled to be true")
	}
	if len(vm.GetBootLogging().GetLoggingServiceIds()) != 1 {
		t.Fatalf("expected one logging service id, got %d", len(vm.GetBootLogging().GetLoggingServiceIds()))
	}
}
