// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package mock_evidence

import (
	"strings"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/testdata"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock Evidence and Resource
var (
	MockEvidence1 = &evidence.Evidence{
		Id:                   testdata.MockEvidenceID1,
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
		ToolId:               testdata.MockEvidenceToolID1,
		Resource:             nil,
	}
	MockEvidence2 = &evidence.Evidence{
		Id:                   testdata.MockEvidenceID2,
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID2,
		ToolId:               testdata.MockEvidenceToolID2,
		Resource:             nil,
	}
	MockEvidence3 = &evidence.Evidence{
		Id:                   testdata.MockEvidenceID2,
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID2,
		ToolId:               testdata.MockEvidenceToolID2,
		Resource: &ontology.Resource{
			Type: &ontology.Resource_VirtualMachine{
				VirtualMachine: &ontology.VirtualMachine{
					Id:   testdata.MockVirtualMachineID1,
					Name: testdata.MockVirtualMachineName1,
				},
			},
		},
	}
	MockVirtualMachineResource1 = &evidence.Resource{
		Id:                   testdata.MockVirtualMachineID1,
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
		ResourceType:         strings.Join(testdata.MockVirtualMachineTypes, ","),
		ToolId:               testdata.MockEvidenceToolID2,
		Properties:           &anypb.Any{},
	}
	MockVirtualMachineResource2 = &evidence.Resource{
		Id:                   testdata.MockVirtualMachineID2,
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID2,
		ResourceType:         strings.Join(testdata.MockVirtualMachineTypes, ","),
		ToolId:               testdata.MockEvidenceToolID1,
		Properties:           &anypb.Any{},
	}
	MockBlockStorageResource1 = &evidence.Resource{
		Id:                   testdata.MockBlockStorageID1,
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
		ResourceType:         strings.Join(testdata.MockBlockStorageTypes, ","),
		ToolId:               testdata.MockEvidenceToolID1,
		Properties:           &anypb.Any{},
	}
	MockBlockStorageResource2 = &evidence.Resource{
		Id:                   testdata.MockBlockStorageID2,
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID2,
		ResourceType:         strings.Join(testdata.MockBlockStorageTypes, ","),
		ToolId:               testdata.MockEvidenceToolID1,
		Properties:           &anypb.Any{},
	}
)
