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

	"confirmate.io/collectors/cloud/internal/constants"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"google.golang.org/protobuf/types/known/durationpb"
)

func Test_azureCollector_collectBackupVaults(t *testing.T) {
	type fields struct {
		azureCollector *azureCollector
	}
	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[*azureCollector]
		wantErr assert.WantErr
	}{
		{
			name: "Backup vaults already collected",
			fields: fields{
				azureCollector: &azureCollector{
					backupMap: map[string]*backup{
						"testBackup": {
							backup: make(map[string][]*ontology.Backup),
						},
					},
				},
			},
			want:    assert.NotNil[*azureCollector],
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: storage account",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			want: func(t *testing.T, got *azureCollector, msgAndArgs ...any) bool {
				want := []*ontology.Backup{
					{
						RetentionPeriod: durationpb.New(Duration7Days),
						Enabled:         true,
						StorageId:       new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.DataProtection/backupVaults/backupAccount1"),
						TransportEncryption: &ontology.TransportEncryption{
							Enforced:        true,
							Enabled:         true,
							ProtocolVersion: 1.2,
							Protocol:        constants.TLS,
						},
					},
				}

				return assert.Equal(t, want, got.backupMap[DataSourceTypeStorageAccountObject].backup["/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.Storage/storageAccounts/account1"])
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: compute disk",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			want: func(t *testing.T, got *azureCollector, msgAndArgs ...any) bool {
				want := []*ontology.Backup{
					{
						RetentionPeriod: durationpb.New(Duration30Days),
						Enabled:         true,
						StorageId:       new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.DataProtection/backupVaults/backupAccount1"),
						TransportEncryption: &ontology.TransportEncryption{
							Enforced:        true,
							Enabled:         true,
							ProtocolVersion: 1.2,
							Protocol:        constants.TLS,
						},
					},
				}

				return assert.Equal(t, want, got.backupMap[DataSourceTypeDisc].backup["/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.Compute/disks/disk1"])
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.azureCollector

			err := d.collectBackupVaults()
			tt.wantErr(t, err)
			tt.want(t, d)
		})
	}
}

func Test_azureCollector_collectBackupInstances(t *testing.T) {
	type fields struct {
		azureCollector       *azureCollector
		clientBackupInstance bool
	}
	type args struct {
		resourceGroup string
		vaultName     string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*armdataprotection.BackupInstanceResource
		wantErr assert.WantErr
	}{
		{
			name: "Input empty",
			args: args{},
			want: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "missing resource group and/or vault name")
			},
		},
		{
			name: "defenderClient not set",
			fields: fields{
				azureCollector:       NewMockAzureCollector(nil),
				clientBackupInstance: true,
			},
			args: args{
				resourceGroup: "res1",
				vaultName:     "backupAccount1",
			},
			want: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not get next page: GET")
			},
		},
		{
			name: "Happy path",
			fields: fields{
				azureCollector:       NewMockAzureCollector(newMockSender()),
				clientBackupInstance: true,
			},
			args: args{
				resourceGroup: "res1",
				vaultName:     "backupAccount1",
			},
			wantErr: assert.NoError,
			want: []*armdataprotection.BackupInstanceResource{
				{
					ID:   new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.DataProtection/backupVaults/backupAccount1/backupInstances/account1-account1-22222222-2222-2222-2222-222222222222"),
					Name: new("account1-account1-22222222-2222-2222-2222-222222222222"),
					Properties: &armdataprotection.BackupInstance{
						DataSourceInfo: &armdataprotection.Datasource{
							ResourceID:     new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.Storage/storageAccounts/account1"),
							DatasourceType: new("Microsoft.Storage/storageAccounts/blobServices"),
						},
						PolicyInfo: &armdataprotection.PolicyInfo{
							PolicyID: new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.DataProtection/backupVaults/backupAccount1/backupPolicies/backupPolicyContainer"),
						},
					},
				},
				{
					ID:   new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.DataProtection/backupVaults/backupAccount1/backupInstances/disk1-disk1-22222222-2222-2222-2222-222222222222"),
					Name: new("disk1-disk1-22222222-2222-2222-2222-222222222222"),
					Properties: &armdataprotection.BackupInstance{
						DataSourceInfo: &armdataprotection.Datasource{
							ResourceID:     new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.Compute/disks/disk1"),
							DatasourceType: new("Microsoft.Compute/disks"),
						},
						PolicyInfo: &armdataprotection.PolicyInfo{
							PolicyID: new("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.DataProtection/backupVaults/backupAccount1/backupPolicies/backupPolicyDisk"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.azureCollector

			if tt.fields.clientBackupInstance {
				// initialize backup instances client
				_ = d.initBackupInstancesClient()
			}
			got, err := d.collectBackupInstances(tt.args.resourceGroup, tt.args.vaultName)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_backupsEmptyCheck(t *testing.T) {
	type args struct {
		backups []*ontology.Backup
	}
	tests := []struct {
		name string
		args args
		want []*ontology.Backup
	}{
		{
			name: "Happy path",
			args: args{
				backups: []*ontology.Backup{
					{
						Enabled:         true,
						Interval:        durationpb.New(90 * time.Hour * 24),
						RetentionPeriod: durationpb.New(100 * time.Hour * 24),
					},
				},
			},
			want: []*ontology.Backup{
				{
					Enabled:         true,
					Interval:        durationpb.New(90 * time.Hour * 24),
					RetentionPeriod: durationpb.New(100 * time.Hour * 24),
				},
			},
		},
		{
			name: "Happy path: empty input",
			args: args{},
			want: []*ontology.Backup{
				{
					Enabled:         false,
					RetentionPeriod: nil,
					Interval:        nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := backupsEmptyCheck(tt.args.backups)
			assert.Equal(t, tt.want, got)
		})
	}
}
