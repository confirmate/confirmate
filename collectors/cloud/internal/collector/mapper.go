package collector

import (
	"fmt"
	"time"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// MapVMToEvidence maps a VM DTO to a minimal valid evidence record.
func MapVMToEvidence(vm AzureVirtualMachine, cfg Config, id string, now time.Time) (ev *evidence.Evidence, err error) {
	if vm.ID == "" {
		return nil, fmt.Errorf("vm id is required")
	}
	if vm.Name == "" {
		return nil, fmt.Errorf("vm name is required")
	}
	if cfg.TargetOfEvaluationID == "" {
		return nil, fmt.Errorf("target of evaluation ID is required")
	}
	if cfg.ToolID == "" {
		return nil, fmt.Errorf("tool ID is required")
	}
	if id == "" {
		return nil, fmt.Errorf("evidence id is required")
	}

	ev = &evidence.Evidence{
		Id:                   id,
		Timestamp:            timestamppb.New(now.UTC()),
		TargetOfEvaluationId: cfg.TargetOfEvaluationID,
		ToolId:               cfg.ToolID,
		Resource: &ontology.Resource{
			Type: &ontology.Resource_VirtualMachine{
				VirtualMachine: mapVM(vm),
			},
		},
	}

	return ev, nil
}

func mapVM(vm AzureVirtualMachine) *ontology.VirtualMachine {
	resource := &ontology.VirtualMachine{
		Id:     vm.ID,
		Name:   vm.Name,
		Labels: vm.Tags,
	}

	if vm.BootLoggingEnabled != nil {
		resource.BootLogging = &ontology.BootLogging{
			Enabled: *vm.BootLoggingEnabled,
		}

		if vm.BootLoggingStoreURI != "" {
			resource.BootLogging.LoggingServiceIds = []string{vm.BootLoggingStoreURI}
		}
	}

	return resource
}
