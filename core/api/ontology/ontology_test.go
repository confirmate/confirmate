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

package ontology

import (
	"testing"
	"time"

	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestResourceTypes(t *testing.T) {
	tests := []struct {
		name string
		r    IsResource
		want []string
	}{
		{
			name: "happy path",
			r:    &VirtualMachine{},
			want: []string{"VirtualMachine", "Compute", "Infrastructure", "Resource"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResourceTypes(tt.r)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRelated(t *testing.T) {
	tests := []struct {
		name string
		r    IsResource
		want []Relationship
	}{
		{
			name: "happy path",
			r: &VirtualMachine{
				Id:              "some-id",
				Name:            "some-name",
				ParentId:        util.Ref("some-parent-id"),
				BlockStorageIds: []string{"some-storage-id"},
			},
			want: []Relationship{
				{
					Property: "block_storage",
					Value:    "some-storage-id",
				},
				{
					Property: "parent",
					Value:    "some-parent-id",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Related(tt.r)

			assert.Equal(t, len(tt.want), len(got))
			for _, rel := range tt.want {
				assert.Contains(t, got, rel)
			}
		})
	}
}

func TestResourceMap(t *testing.T) {
	tests := []struct {
		name      string
		r         IsResource
		wantProps assert.Want[map[string]any]
		wantErr   assert.WantErr
	}{
		{
			name: "happy path",
			r: &VirtualMachine{
				Id:           "my-id",
				Name:         "My VM",
				CreationTime: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				AutomaticUpdates: &AutomaticUpdates{
					Interval: durationpb.New(time.Hour * 24 * 2),
				},
			},
			wantProps: func(t *testing.T, got map[string]any, msgAndArgs ...any) bool {
				want := map[string]any{
					"activityLogging":            nil,
					"blockStorageIds":            []any{},
					"bootLogging":                nil,
					"creationTime":               "2024-01-01T00:00:00Z",
					"encryptionInUse":            nil,
					"geoLocation":                nil,
					"id":                         "my-id",
					"internetAccessibleEndpoint": false,
					"labels":                     map[string]any{},
					"name":                       "My VM",
					"description":                "",
					"networkInterfaceIds":        []any{},
					"malwareProtection":          nil,
					"osLogging":                  nil,
					"loggings":                   []any{},
					"raw":                        "",
					"redundancies":               []any{},
					"remoteAttestation":          nil,
					"resourceLogging":            nil,
					"automaticUpdates": map[string]any{
						"enabled":      false,
						"interval":     "172800s",
						"securityOnly": false,
					},
					"type":            []string{"VirtualMachine", "Compute", "Infrastructure", "Resource"},
					"usageStatistics": nil,
				}

				return assert.Equal(t, want, got)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProps, err := ResourceMap(tt.r)

			tt.wantErr(t, err)
			tt.wantProps(t, gotProps)
		})
	}
}

func TestResourceIDs(t *testing.T) {
	tests := []struct {
		name string
		r    []IsResource
		want []string
	}{
		{
			name: "empty input",
			r:    nil,
			want: []string{},
		},
		{
			name: "happy path",
			r: []IsResource{
				&Account{Id: "test"},
				&Account{Id: "test2"},
			},
			want: []string{"test", "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ResourceIDs(tt.r))
		})
	}
}

func TestProtoResource(t *testing.T) {
	tests := []struct {
		name string
		r    IsResource
		want *Resource
	}{
		{
			name: "happy path",
			r: &VirtualMachine{
				Id: "vm-1",
			},
			want: &Resource{
				Type: &Resource_VirtualMachine{
					VirtualMachine: &VirtualMachine{
						Id: "vm-1",
					},
				},
			},
		},
		{
			name: "nil input",
			r:    nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ProtoResource(tt.r))
		})
	}
}

func TestListResourceTypes(t *testing.T) {
	tests := []struct {
		name string
		want assert.Want[[]string]
	}{
		{
			name: "happy path",
			want: func(t *testing.T, got []string, msgAndArgs ...any) bool {
				return assert.NotEmpty(t, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want(t, ListResourceTypes())
		})
	}
}

func TestResourceJSONRoundTrip(t *testing.T) {
	want := &Resource{
		Type: &Resource_VirtualMachine{
			VirtualMachine: &VirtualMachine{
				Id:   "vm-1",
				Name: "vm-name",
			},
		},
	}

	data, err := want.MarshalJSON()
	assert.NoError(t, err)

	got := new(Resource)
	err = got.UnmarshalJSON(data)
	assert.NoError(t, err)

	assert.Equal(t, want, got)
}
