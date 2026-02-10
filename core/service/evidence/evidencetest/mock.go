package evidencetest

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
)

var (
	MockEvidence1 = &evidence.Evidence{
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
	// MockEvidence2SameResourceAs1 is a second evidence (new ID) that references the same resource as MockEvidence1 to
	// test resource upsert behavior (database)
	MockEvidence2SameResourceAs1 = &evidence.Evidence{
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
	MockEvidenceNoResource = &evidence.Evidence{
		Id:                   uuid.NewString(),
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: uuid.NewString(),
		ToolId:               "MockTool1",
		Resource:             nil,
	}
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
	MockResourceListA = &evidence.Resource{
		Id:                   "vm-1",
		ResourceType:         "virtual_machine",
		TargetOfEvaluationId: "11111111-1111-1111-1111-111111111111",
		ToolId:               "tool-a",
		Properties:           nil,
	}
	MockResourceListB = &evidence.Resource{
		Id:                   "app-1",
		ResourceType:         "application",
		TargetOfEvaluationId: "22222222-2222-2222-2222-222222222222",
		ToolId:               "tool-a",
		Properties:           nil,
	}
	MockResourceListC = &evidence.Resource{
		Id:                   "vm-2",
		ResourceType:         "virtual_machine",
		TargetOfEvaluationId: "11111111-1111-1111-1111-111111111111",
		ToolId:               "tool-b",
		Properties:           nil,
	}
)
