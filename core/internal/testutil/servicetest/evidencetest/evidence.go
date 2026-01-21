package evidencetest

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/persistence"
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
)
var InitDBWithEvidence = func(db persistence.DB) {
	err := db.Create(MockEvidence1)
	if err != nil {
		panic(err)
	}
}
