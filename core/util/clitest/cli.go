package clitest

import (
	"errors"
	"os"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/testdata"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var MockAssessmentResult1 = &assessment.AssessmentResult{
	Id:                   testdata.MockAssessmentResult1ID,
	CreatedAt:            timestamppb.New(time.Unix(1, 0)),
	TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
	MetricId:             testdata.MockMetricID1,
	Compliant:            true,
	EvidenceId:           testdata.MockEvidenceID1,
	ResourceId:           testdata.MockVirtualMachineID1,
	ResourceTypes:        testdata.MockVirtualMachineTypes,
	ComplianceComment:    assessment.DefaultCompliantMessage,
	MetricConfiguration: &assessment.MetricConfiguration{
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            true,
		MetricId:             testdata.MockMetricID1,
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
	},
	ToolId:           util.Ref(assessment.AssessmentToolId),
	HistoryUpdatedAt: timestamppb.New(time.Unix(1, 0)),
	History: []*assessment.Record{
		{
			EvidenceId:         testdata.MockEvidenceID1,
			EvidenceRecordedAt: timestamppb.New(time.Unix(1, 0)),
		},
	},
}

var (
	MockEvidence1 = &evidence.Evidence{
		Id:                   testdata.MockEvidenceID1,
		Timestamp:            timestamppb.New(time.Unix(1, 0)),
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
		ToolId:               testdata.MockEvidenceToolID1,
		Resource: &ontology.Resource{
			Type: &ontology.Resource_VirtualMachine{
				VirtualMachine: &ontology.VirtualMachine{
					Id:           testdata.MockVirtualMachineID1,
					Name:         testdata.MockVirtualMachineName1,
					Description:  "Mock evidence for Virtual Machine",
					CreationTime: timestamppb.New(time.Unix(1, 0)),
					AutomaticUpdates: &ontology.AutomaticUpdates{
						Enabled: true,
					},
					BlockStorageIds: []string{testdata.MockVirtualMachineID2},
				},
			},
		},
	}

	MockEvidence2 = &evidence.Evidence{
		Id:                   testdata.MockEvidenceID2,
		Timestamp:            timestamppb.New(time.Unix(1, 0)),
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
		ToolId:               testdata.MockEvidenceToolID1,
		Resource: &ontology.Resource{
			Type: &ontology.Resource_BlockStorage{
				BlockStorage: &ontology.BlockStorage{
					Id:           testdata.MockBlockStorageID1,
					Name:         testdata.MockBlockStorageName1,
					Description:  "Mock evidence for Block Storage",
					CreationTime: timestamppb.New(time.Unix(1, 0)),
				},
			},
		},
	}
)

// AutoChdir automatically guesses if we need to change the current working directory
// so that we can find the policies folder
func AutoChdir() {
	var (
		err error
	)

	// Check, if we can find the core folder
	_, err = os.Stat("policies")
	if errors.Is(err, os.ErrNotExist) {
		// Try again one level deeper
		err = os.Chdir("..")
		if err != nil {
			panic(err)
		}

		AutoChdir()
	}
}
