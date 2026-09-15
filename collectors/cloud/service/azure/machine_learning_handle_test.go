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
	"time"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_handleMLWorkspace(t *testing.T) {
	creationTime := time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)
	id := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace"
	parent := "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/rg1"
	storage := "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/rg1/providers/microsoft.storage/storageaccounts/account1"
	applicationInsights := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.insights/components/appInsights1"
	keyVault := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.Keyvault/vaults/keyVault1"

	type fields struct {
		d *azureCollector
	}
	type args struct {
		value       *armmachinelearning.Workspace
		computeList []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[ontology.IsResource]
		wantErr assert.WantErr
	}{
		{
			name: "Happy path",
			fields: fields{
				d: &azureCollector{},
			},
			args: args{
				value: &armmachinelearning.Workspace{
					Name: new("mlWorkspace"),
					ID:   new(id),
					SystemData: &armmachinelearning.SystemData{
						CreatedAt: new(creationTime),
					},
					Tags:     map[string]*string{"tag1": new("tag1"), "tag2": new("tag2")},
					Location: new("westeurope"),
					Properties: &armmachinelearning.WorkspaceProperties{
						PublicNetworkAccess: new(armmachinelearning.PublicNetworkAccessEnabled),
						ApplicationInsights: new(applicationInsights),
						Encryption: &armmachinelearning.EncryptionProperty{
							Status: new(armmachinelearning.EncryptionStatusEnabled),
							KeyVaultProperties: &armmachinelearning.KeyVaultProperties{
								KeyVaultArmID: new(keyVault),
							},
						},
						StorageAccount: new(storage),
					},
				},
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				got1 := got.(*ontology.MachineLearningService)

				want := &ontology.MachineLearningService{
					Id:                         resourceID(new(id)),
					Name:                       "mlWorkspace",
					CreationTime:               timestamppb.New(creationTime),
					GeoLocation:                &ontology.GeoLocation{Region: "westeurope"},
					Labels:                     map[string]string{"tag1": "tag1", "tag2": "tag2"},
					ParentId:                   new(parent),
					InternetAccessibleEndpoint: true,
					StorageIds:                 []string{storage},
					ComputeIds:                 []string{},
					Loggings: []*ontology.Logging{
						{
							Type: &ontology.Logging_ResourceLogging{
								ResourceLogging: &ontology.ResourceLogging{
									Enabled:           true,
									LoggingServiceIds: []string{resourceID(new(applicationInsights))},
								},
							},
						},
					},
				}

				assert.NotEmpty(t, got1.Raw)
				got1.Raw = ""

				return assert.Equal(t, want, got1)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.fields.d.handleMLWorkspace(tt.args.value, tt.args.computeList)

			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}

func Test_azureCollector_handleMLCompute(t *testing.T) {
	type fields struct {
		isAuthorized       bool
		sub                *armsubscription.Subscription
		cred               azcore.TokenCredential
		rg                 *string
		clientOptions      arm.ClientOptions
		clients            clients
		ctID               string
		backupMap          map[string]*backup
		defenderProperties map[string]*defenderProperties
	}
	type args struct {
		value       *armmachinelearning.ComputeResource
		workspaceID *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[ontology.IsResource]
		wantErr assert.WantErr
	}{
		{
			name: "Happy path: ComputeInstance",
			args: args{
				value: &armmachinelearning.ComputeResource{
					Name: new("compute1"),
					ID:   new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace/computes/compute1"),
					SystemData: &armmachinelearning.SystemData{
						CreatedAt: new(time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)),
					},
					Tags:     map[string]*string{"tag1": new("tag1"), "tag2": new("tag2")},
					Location: new("westeurope"),
					Properties: &armmachinelearning.ComputeInstance{
						ComputeLocation: new("westeurope"),
					},
				},
				workspaceID: new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace"),
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				got1 := got.(*ontology.Container)

				want := &ontology.Container{
					Id:                  resourceID(new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace/computes/compute1")),
					Name:                "compute1",
					CreationTime:        timestamppb.New(time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)),
					GeoLocation:         &ontology.GeoLocation{Region: "westeurope"},
					Labels:              map[string]string{"tag1": "tag1", "tag2": "tag2"},
					ParentId:            resourceIDPointer(new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace")),
					NetworkInterfaceIds: []string{},
				}

				assert.NotEmpty(t, got1.Raw)
				got1.Raw = ""

				return assert.Equal(t, want, got1)
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: VirtualMachine",
			args: args{
				value: &armmachinelearning.ComputeResource{
					Name: new("compute1"),
					ID:   new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace/computes/compute1"),
					SystemData: &armmachinelearning.SystemData{
						CreatedAt: new(time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)),
					},
					Tags:     map[string]*string{"tag1": new("tag1"), "tag2": new("tag2")},
					Location: new("westeurope"),
					Properties: &armmachinelearning.VirtualMachine{
						ComputeLocation: new("westeurope"),
					},
				},
				workspaceID: new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace"),
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				got1 := got.(*ontology.VirtualMachine)

				want := &ontology.VirtualMachine{
					Id:                  resourceID(new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace/computes/compute1")),
					Name:                "compute1",
					CreationTime:        timestamppb.New(time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)),
					GeoLocation:         &ontology.GeoLocation{Region: "westeurope"},
					Labels:              map[string]string{"tag1": "tag1", "tag2": "tag2"},
					ParentId:            resourceIDPointer(new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace")),
					NetworkInterfaceIds: []string{},
					MalwareProtection:   &ontology.MalwareProtection{},
				}

				assert.NotEmpty(t, got1.Raw)
				got1.Raw = ""

				return assert.Equal(t, want, got1)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &azureCollector{
				isAuthorized:       tt.fields.isAuthorized,
				sub:                tt.fields.sub,
				cred:               tt.fields.cred,
				rg:                 tt.fields.rg,
				clientOptions:      tt.fields.clientOptions,
				clients:            tt.fields.clients,
				ctID:               tt.fields.ctID,
				backupMap:          tt.fields.backupMap,
				defenderProperties: tt.fields.defenderProperties,
			}
			got, err := d.handleMLCompute(tt.args.value, tt.args.workspaceID)

			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
