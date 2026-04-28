// Copyright 2016-2026 Fraunhofer AISEC
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

package azure

import (
	"testing"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
)

func Test_azureCollector_collectMLWorkspaces(t *testing.T) {
	type fields struct {
		azureCollector *azureCollector
	}
	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[[]ontology.IsResource]
		wantErr assert.WantErr
	}{
		{
			name: "Error list pages",
			fields: fields{
				azureCollector: NewMockAzureCollector(nil),
			},
			want: assert.Nil[[]ontology.IsResource],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, ErrSubscriptionNotFound.Error())
			},
		},
		{
			name: "Happy path",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				assert.Equal(t, got[0].GetName(), "compute1")
				assert.Equal(t, got[1].GetName(), "mlWorkspace")
				return assert.Equal(t, 2, len(got))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.azureCollector

			got, err := d.collectMLWorkspaces()

			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}

func Test_azureCollector_collectMLCompute(t *testing.T) {
	type fields struct {
		azureCollector *azureCollector
	}
	type args struct {
		rg        string
		workspace *armmachinelearning.Workspace
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[[]ontology.IsResource]
		wantErr assert.WantErr
	}{
		{
			name: "Error list pages",
			fields: fields{
				azureCollector: NewMockAzureCollector(nil),
			},
			args: args{
				rg: "rg",
				workspace: &armmachinelearning.Workspace{
					Name: new("mlWorkspace"),
				},
			},
			want: assert.Nil[[]ontology.IsResource],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, ErrSubscriptionNotFound.Error())
			},
		},
		{
			name: "Happy path",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			args: args{
				rg: "rg1",
				workspace: &armmachinelearning.Workspace{
					Name: new("mlWorkspace"),
				},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				assert.Equal(t, 1, len(got))

				_, ok := got[0].(*ontology.VirtualMachine)
				return assert.True(t, ok)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.azureCollector

			got, err := d.collectMLCompute(tt.args.rg, tt.args.workspace)

			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
