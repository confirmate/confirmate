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

package evidence

import (
	"testing"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/service/evidence/evidencetest"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/prototest"
	anypb "google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEvidence_GetOntologyResource(t *testing.T) {
	type fields struct {
		Id                             string
		Timestamp                      *timestamppb.Timestamp
		TargetOfEvaluationId           string
		ToolId                         string
		Resource                       *ontology.Resource
		ExperimentalRelatedResourceIds []string
	}
	tests := []struct {
		name   string
		fields fields
		want   ontology.IsResource
	}{
		{
			name: "happy path",
			fields: fields{
				Resource: &ontology.Resource{
					Type: &ontology.Resource_VirtualMachine{
						VirtualMachine: &ontology.VirtualMachine{
							Id: "vm-1",
						},
					},
				},
			},
			want: &ontology.VirtualMachine{
				Id: "vm-1",
			},
		},
		{
			name: "resource is nil",
			fields: fields{
				Resource: nil,
			},
			want: nil,
		},
		{
			name: "resource is empty",
			fields: fields{
				Resource: &ontology.Resource{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &Evidence{
				Id:                             tt.fields.Id,
				Timestamp:                      tt.fields.Timestamp,
				TargetOfEvaluationId:           tt.fields.TargetOfEvaluationId,
				ToolId:                         tt.fields.ToolId,
				Resource:                       tt.fields.Resource,
				ExperimentalRelatedResourceIds: tt.fields.ExperimentalRelatedResourceIds,
			}

			assert.Equal(t, tt.want, ev.GetOntologyResource())
		})
	}
}

func TestResource_ToOntologyResource(t *testing.T) {
	type fields struct {
		Id                   string
		TargetOfEvaluationId string
		ResourceType         string
		Properties           *anypb.Any
	}
	tests := []struct {
		name    string
		fields  fields
		want    ontology.IsResource
		wantErr assert.WantErr
	}{
		{
			name: "happy path VM",
			fields: fields{
				Id:                   "vm1",
				TargetOfEvaluationId: "target1",
				ResourceType:         "VirtualMachine",
				Properties: prototest.NewAny(t, &ontology.VirtualMachine{
					Id:              "vm1",
					BlockStorageIds: []string{"bs1"},
				}),
			},
			want: &ontology.VirtualMachine{
				Id:              "vm1",
				BlockStorageIds: []string{"bs1"},
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "not an ontology resource",
			fields: fields{
				Id:                   "vm1",
				TargetOfEvaluationId: "target1",
				ResourceType:         "Something",
				Properties:           prototest.NewAny(t, &emptypb.Empty{}),
			},
			want: nil,
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, ontology.ErrNotOntologyResource.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{
				Id:                   tt.fields.Id,
				TargetOfEvaluationId: tt.fields.TargetOfEvaluationId,
				ResourceType:         tt.fields.ResourceType,
				Properties:           tt.fields.Properties,
			}
			got, err := r.ToOntologyResource()

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToEvidenceResource(t *testing.T) {
	type args struct {
		resource    ontology.IsResource
		ctID        string
		collectorID string
	}
	tests := []struct {
		name    string
		args    args
		want    *Resource
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				resource: &ontology.BlockStorage{
					Id:   "my-block-storage",
					Name: "My Block Storage",
					Backups: []*ontology.Backup{
						{
							Enabled:   true,
							StorageId: util.Ref("my-offsite-backup-id"),
						},
					},
				},
				ctID:        evidencetest.MockTargetOfEvaluationID1,
				collectorID: evidencetest.MockEvidenceToolID1,
			},
			want: &Resource{
				Id:                   "my-block-storage",
				TargetOfEvaluationId: evidencetest.MockTargetOfEvaluationID1,
				ToolId:               evidencetest.MockEvidenceToolID1,
				ResourceType:         "BlockStorage,Storage,Infrastructure,Resource",
				Properties: prototest.NewAny(t, &ontology.BlockStorage{
					Id:   "my-block-storage",
					Name: "My Block Storage",
					Backups: []*ontology.Backup{
						{
							Enabled:   true,
							StorageId: util.Ref("my-offsite-backup-id"),
						},
					},
				}),
			},
			wantErr: assert.Nil[error],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, err := ToEvidenceResource(tt.args.resource, tt.args.ctID, tt.args.collectorID)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, gotR)
		})
	}
}
