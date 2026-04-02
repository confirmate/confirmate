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

	"google.golang.org/protobuf/types/known/timestamppb"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
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

func TestToResourceSnapshot(t *testing.T) {
	type args struct {
		resource    ontology.IsResource
		ctID        string
		collectorID string
	}
	tests := []struct {
		name    string
		args    args
		want    *ResourceSnapshot
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
				ctID:        "test-toe-id",
				collectorID: "test-collector-id",
			},
			want: &ResourceSnapshot{
				Id:                   "my-block-storage",
				TargetOfEvaluationId: "test-toe-id",
				ToolId:               "test-collector-id",
				ResourceType:         "BlockStorage,Storage,Infrastructure,Resource",
				Resource: &ontology.Resource{
					Type: &ontology.Resource_BlockStorage{
						BlockStorage: &ontology.BlockStorage{
							Id:   "my-block-storage",
							Name: "My Block Storage",
							Backups: []*ontology.Backup{
								{
									Enabled:   true,
									StorageId: util.Ref("my-offsite-backup-id"),
								},
							},
						},
					},
				},
			},
			wantErr: assert.Nil[error],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, err := ToResourceSnapshot(tt.args.resource, tt.args.ctID, tt.args.collectorID)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, gotR)
		})
	}
}
