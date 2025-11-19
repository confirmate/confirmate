package azure

import (
	"testing"
	"time"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
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
		d *azureDiscovery
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
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			fields: fields{
				d: &azureDiscovery{},
			},
			args: args{
				value: &armmachinelearning.Workspace{
					Name: util.Ref("mlWorkspace"),
					ID:   util.Ref(id),
					SystemData: &armmachinelearning.SystemData{
						CreatedAt: util.Ref(creationTime),
					},
					Tags:     map[string]*string{"tag1": util.Ref("tag1"), "tag2": util.Ref("tag2")},
					Location: util.Ref("westeurope"),
					Properties: &armmachinelearning.WorkspaceProperties{
						PublicNetworkAccess: util.Ref(armmachinelearning.PublicNetworkAccessEnabled),
						ApplicationInsights: util.Ref(applicationInsights),
						Encryption: &armmachinelearning.EncryptionProperty{
							Status: util.Ref(armmachinelearning.EncryptionStatusEnabled),
							KeyVaultProperties: &armmachinelearning.KeyVaultProperties{
								KeyVaultArmID: util.Ref(keyVault),
							},
						},
						StorageAccount: util.Ref(storage),
					},
				},
			},
			want: func(t *testing.T, got ontology.IsResource) bool {
				got1 := got.(*ontology.MachineLearningService)

				want := &ontology.MachineLearningService{
					Id:                         resourceID(util.Ref(id)),
					Name:                       "mlWorkspace",
					CreationTime:               timestamppb.New(creationTime),
					GeoLocation:                &ontology.GeoLocation{Region: "westeurope"},
					Labels:                     map[string]string{"tag1": "tag1", "tag2": "tag2"},
					ParentId:                   util.Ref(parent),
					InternetAccessibleEndpoint: true,
					StorageIds:                 []string{storage},
					ComputeIds:                 []string{},
					Loggings: []*ontology.Logging{
						{
							Type: &ontology.Logging_ResourceLogging{
								ResourceLogging: &ontology.ResourceLogging{
									Enabled:           true,
									LoggingServiceIds: []string{resourceID(util.Ref(applicationInsights))},
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

func Test_azureDiscovery_handleMLCompute(t *testing.T) {
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
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path: ComputeInstance",
			args: args{
				value: &armmachinelearning.ComputeResource{
					Name: util.Ref("compute1"),
					ID:   util.Ref("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace/computes/compute1"),
					SystemData: &armmachinelearning.SystemData{
						CreatedAt: util.Ref(time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)),
					},
					Tags:     map[string]*string{"tag1": util.Ref("tag1"), "tag2": util.Ref("tag2")},
					Location: util.Ref("westeurope"),
					Properties: &armmachinelearning.ComputeInstance{
						ComputeLocation: util.Ref("westeurope"),
					},
				},
				workspaceID: util.Ref("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace"),
			},
			want: func(t *testing.T, got ontology.IsResource) bool {
				got1 := got.(*ontology.Container)

				want := &ontology.Container{
					Id:                  resourceID(util.Ref("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace/computes/compute1")),
					Name:                "compute1",
					CreationTime:        timestamppb.New(time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)),
					GeoLocation:         &ontology.GeoLocation{Region: "westeurope"},
					Labels:              map[string]string{"tag1": "tag1", "tag2": "tag2"},
					ParentId:            resourceIDPointer(util.Ref("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace")),
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
					Name: util.Ref("compute1"),
					ID:   util.Ref("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace/computes/compute1"),
					SystemData: &armmachinelearning.SystemData{
						CreatedAt: util.Ref(time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)),
					},
					Tags:     map[string]*string{"tag1": util.Ref("tag1"), "tag2": util.Ref("tag2")},
					Location: util.Ref("westeurope"),
					Properties: &armmachinelearning.VirtualMachine{
						ComputeLocation: util.Ref("westeurope"),
					},
				},
				workspaceID: util.Ref("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace"),
			},
			want: func(t *testing.T, got ontology.IsResource) bool {
				got1 := got.(*ontology.VirtualMachine)

				want := &ontology.VirtualMachine{
					Id:                  resourceID(util.Ref("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace/computes/compute1")),
					Name:                "compute1",
					CreationTime:        timestamppb.New(time.Date(2017, 05, 24, 13, 28, 53, 4540398, time.UTC)),
					GeoLocation:         &ontology.GeoLocation{Region: "westeurope"},
					Labels:              map[string]string{"tag1": "tag1", "tag2": "tag2"},
					ParentId:            resourceIDPointer(util.Ref("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.MachineLearningServices/workspaces/mlWorkspace")),
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
			d := &azureDiscovery{
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
