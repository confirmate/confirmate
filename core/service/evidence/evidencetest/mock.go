package evidencetest

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
)

var (
	// MockEvidenceWithVMResource is a generic, well-formed evidence used as a default “happy path” evidence in tests.
	//	// It includes an ontology VM resource and uses random IDs (uuid.NewString()).
	MockEvidenceWithVMResource = &evidence.Evidence{
		Id:                   uuid.NewString(),
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: uuid.NewString(),
		ToolId:               "MockTool1",
		Resource: &ontology.Resource{Type: &ontology.Resource_VirtualMachine{
			VirtualMachine: &ontology.VirtualMachine{
				Id:   "mock-id-1",
				Name: "my-vm",
			},
		}},
	}
	// MockEvidenceWithVMResource2 is a second evidence (with a new Evidence ID) that references the same ontology resource
	// as [MockEvidenceWithVMResource]. Use it to test “upsert”/deduplication behavior for resources in persistence.
	MockEvidenceWithVMResource2 = &evidence.Evidence{
		Id:                   uuid.NewString(),
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: uuid.NewString(),
		ToolId:               "MockTool1",
		Resource: &ontology.Resource{Type: &ontology.Resource_VirtualMachine{
			VirtualMachine: &ontology.VirtualMachine{
				Id:   "mock-id-1",
				Name: "my-vm",
			},
		}},
	}

	// MockEvidenceWithoutResource is an evidence intentionally missing the resource (Resource == nil).
	//	// Use it to test validation/error paths and nil-handling in store/convert logic.
	MockEvidenceWithoutResource = &evidence.Evidence{
		Id:                   uuid.NewString(),
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: uuid.NewString(),
		ToolId:               "MockTool1",
		Resource:             nil,
	}

	// MockEvidenceListA is a deterministic evidence fixture for list/filter tests.
	// It represents tool "tool-a" and a fixed TargetOfEvaluationId (ToE) and VM resource identity.
	MockEvidenceListA = &evidence.Evidence{
		Id:                   "00000000-0000-0000-0000-000000000001",
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: "11111111-1111-1111-1111-111111111111",
		ToolId:               "tool-a",
		Resource: &ontology.Resource{Type: &ontology.Resource_VirtualMachine{
			VirtualMachine: &ontology.VirtualMachine{
				Id:   "vm-1",
				Name: "vm-vm-1",
			},
		}},
	}
	// MockEvidenceListB is a deterministic evidence fixture for list/filter tests.
	// Compared to [MockEvidenceListA] it varies the ToE and VM identity, while keeping tool "tool-a".
	MockEvidenceListB = &evidence.Evidence{
		Id:                   "00000000-0000-0000-0000-000000000002",
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: "22222222-2222-2222-2222-222222222222",
		ToolId:               "tool-a",
		Resource: &ontology.Resource{Type: &ontology.Resource_VirtualMachine{
			VirtualMachine: &ontology.VirtualMachine{
				Id:   "vm-2",
				Name: "vm-vm-2",
			},
		}},
	}
	// MockEvidenceListC is a deterministic evidence fixture for list/filter tests.
	// Compared to [MockEvidenceListA] it varies the tool (tool-b) while keeping the same ToE as ListA.
	MockEvidenceListC = &evidence.Evidence{
		Id:                   "00000000-0000-0000-0000-000000000003",
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: "11111111-1111-1111-1111-111111111111",
		ToolId:               "tool-b",
		Resource: &ontology.Resource{Type: &ontology.Resource_VirtualMachine{
			VirtualMachine: &ontology.VirtualMachine{
				Id:   "vm-3",
				Name: "vm-vm-3",
			},
		}},
	}
	// MockResourceListA is a deterministic resource fixture for list/filter tests.
	// It corresponds to a VM resource for tool "tool-a" and the same ToE as [MockEvidenceListA].
	MockResourceListA = &evidence.Resource{
		Id:                   "vm-1",
		ResourceType:         "virtual_machine",
		TargetOfEvaluationId: "11111111-1111-1111-1111-111111111111",
		ToolId:               "tool-a",
		Properties:           nil,
	}
	// MockResourceListB is a deterministic resource fixture for list/filter tests.
	// Compared to [MockResourceListA] represents a different resource type ("application") for tool "tool-a" and a
	// different ToE.
	MockResourceListB = &evidence.Resource{
		Id:                   "app-1",
		ResourceType:         "application",
		TargetOfEvaluationId: "22222222-2222-2222-2222-222222222222",
		ToolId:               "tool-a",
		Properties:           nil,
	}
	// MockResourceListC is a deterministic resource fixture for list/filter tests.
	// It represents a VM resource for tool "tool-b" and the same ToE as [MockEvidenceListA]/[MockEvidenceListC].
	MockResourceListC = &evidence.Resource{
		Id:                   "vm-2",
		ResourceType:         "virtual_machine",
		TargetOfEvaluationId: "11111111-1111-1111-1111-111111111111",
		ToolId:               "tool-b",
		Properties:           nil,
	}
)
