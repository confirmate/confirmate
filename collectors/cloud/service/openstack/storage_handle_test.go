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

package openstack

import (
	"net/http"
	"testing"
	"time"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/timestamp"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
	"google.golang.org/protobuf/testing/protocmp"
)

// volumeTypeHandlers returns the handlers of a block storage service with one volume type "enc-type", whose
// encryption is the given JSON fragment.
func volumeTypeHandlers(encryption string) func(mux *http.ServeMux) {
	return func(mux *http.ServeMux) {
		mux.HandleFunc("/types", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"volume_types": [{"id": "enc-type-id", "name": "enc-type"}]}`))
		})
		mux.HandleFunc("/types/enc-type-id/encryption", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{` + encryption + `}`))
		})
	}
}

func Test_openstackCollector_handleBlockStorage(t *testing.T) {
	testTime := time.Date(2000, 01, 20, 9, 20, 12, 123, time.UTC)

	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
		region   string
		domain   *domain
		project  *project
		// handlers registers the fake responses of the block storage service
		handlers func(mux *http.ServeMux)
	}
	type args struct {
		volume *volumes.Volume
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[ontology.IsResource]
		wantErr assert.WantErr
	}{
		{
			name: "Happy path: volume name missing",
			fields: fields{
				region: "test region",
			},
			args: args{
				volume: &volumes.Volume{
					ID: testdata.MockOpenstackVolumeID1,
					// Name:      testdata.MockOpenstackVolumeName1,
					TenantID:  testdata.MockOpenstackVolumeTenantID,
					CreatedAt: testTime,
				},
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				want := &ontology.BlockStorage{
					Id:           new(testdata.MockOpenstackVolumeID1),
					Name:         new(testdata.MockOpenstackVolumeID1),
					CreationTime: timestamp.New(testTime),
					GeoLocation: &ontology.GeoLocation{
						Region: new("test region"),
					},
					ParentId:    new(testdata.MockOpenstackVolumeTenantID),
					Description: new(""),
					AtRestEncryption: &ontology.AtRestEncryption{
						Type: &ontology.AtRestEncryption_DiskEncryption{
							DiskEncryption: &ontology.DiskEncryption{Enabled: new(false)},
						},
					},
					Backups: []*ontology.Backup{
						{
							Enabled: new(false),
						},
					},
				}

				gotNew := got.(*ontology.BlockStorage)
				return assert.NotEmpty(t, gotNew.GetRaw()) && assert.Equal(t, want, gotNew, protocmp.IgnoreFields(&ontology.BlockStorage{}, "raw"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: volume name available",
			fields: fields{
				region: "test region",
			},
			args: args{
				volume: &volumes.Volume{
					ID:        testdata.MockOpenstackVolumeID1,
					Name:      testdata.MockOpenstackVolumeName1,
					TenantID:  testdata.MockOpenstackVolumeTenantID,
					CreatedAt: testTime,
				},
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				want := &ontology.BlockStorage{
					Id:           new(testdata.MockOpenstackVolumeID1),
					Name:         new(testdata.MockOpenstackVolumeName1),
					CreationTime: timestamp.New(testTime),
					GeoLocation: &ontology.GeoLocation{
						Region: new("test region"),
					},
					ParentId:    new(testdata.MockOpenstackVolumeTenantID),
					Description: new(""),
					AtRestEncryption: &ontology.AtRestEncryption{
						Type: &ontology.AtRestEncryption_DiskEncryption{
							DiskEncryption: &ontology.DiskEncryption{Enabled: new(false)},
						},
					},
					Backups: []*ontology.Backup{
						{
							Enabled: new(false),
						},
					},
				}

				gotNew := got.(*ontology.BlockStorage)
				return assert.NotEmpty(t, gotNew.GetRaw()) && assert.Equal(t, want, gotNew, protocmp.IgnoreFields(&ontology.BlockStorage{}, "raw"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: encrypted volume, volume type lookup fails",
			fields: fields{
				region: "test region",
			},
			args: args{
				volume: &volumes.Volume{
					ID:        testdata.MockOpenstackVolumeID1,
					Name:      testdata.MockOpenstackVolumeName1,
					TenantID:  testdata.MockOpenstackVolumeTenantID,
					CreatedAt: testTime,
					Encrypted: true,
				},
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				want := &ontology.BlockStorage{
					Id:           new(testdata.MockOpenstackVolumeID1),
					Name:         new(testdata.MockOpenstackVolumeName1),
					CreationTime: timestamp.New(testTime),
					GeoLocation: &ontology.GeoLocation{
						Region: new("test region"),
					},
					ParentId:    new(testdata.MockOpenstackVolumeTenantID),
					Description: new(""),
					AtRestEncryption: &ontology.AtRestEncryption{
						Type: &ontology.AtRestEncryption_DiskEncryption{
							DiskEncryption: &ontology.DiskEncryption{Enabled: new(true)},
						},
					},
					Backups: []*ontology.Backup{
						{
							Enabled: new(false),
						},
					},
				}

				gotNew := got.(*ontology.BlockStorage)
				return assert.NotEmpty(t, gotNew.GetRaw()) && assert.Equal(t, want, gotNew, protocmp.IgnoreFields(&ontology.BlockStorage{}, "raw"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: encrypted volume type",
			fields: fields{
				region:   "test region",
				handlers: volumeTypeHandlers(`"encryption_id": "enc-id", "cipher": "aes-xts-plain64"`),
			},
			args: args{
				volume: &volumes.Volume{
					ID:         testdata.MockOpenstackVolumeID1,
					Name:       testdata.MockOpenstackVolumeName1,
					TenantID:   testdata.MockOpenstackVolumeTenantID,
					CreatedAt:  testTime,
					VolumeType: "enc-type",
				},
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				return assert.Equal(t, &ontology.AtRestEncryption{
					Type: &ontology.AtRestEncryption_DiskEncryption{
						DiskEncryption: &ontology.DiskEncryption{Enabled: new(true), Algorithm: new("aes-xts-plain64")},
					},
				}, got.(*ontology.BlockStorage).GetAtRestEncryption())
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: unencrypted volume type",
			fields: fields{
				region:   "test region",
				handlers: volumeTypeHandlers(`"encryption_id": ""`),
			},
			args: args{
				volume: &volumes.Volume{
					ID:         testdata.MockOpenstackVolumeID1,
					Name:       testdata.MockOpenstackVolumeName1,
					TenantID:   testdata.MockOpenstackVolumeTenantID,
					CreatedAt:  testTime,
					VolumeType: "enc-type",
				},
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				return assert.Equal(t, &ontology.AtRestEncryption{
					Type: &ontology.AtRestEncryption_DiskEncryption{
						DiskEncryption: &ontology.DiskEncryption{Enabled: new(false)},
					},
				}, got.(*ontology.BlockStorage).GetAtRestEncryption())
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeServer := testhelper.SetupHTTP()
			t.Cleanup(fakeServer.Teardown)

			tt.fields.clients.blockStorageClient = client.ServiceClient(fakeServer)
			if tt.fields.handlers != nil {
				tt.fields.handlers(fakeServer.Mux)
			}

			d := &openstackCollector{
				ctID:     tt.fields.ctID,
				clients:  tt.fields.clients,
				authOpts: tt.fields.authOpts,
				region:   tt.fields.region,
				domain:   tt.fields.domain,
				project:  tt.fields.project,
			}
			got, err := d.handleBlockStorage(tt.args.volume)

			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}

func Test_openstackCollector_handleObjectStorage(t *testing.T) {
	fakeServer := testhelper.SetupHTTP()
	defer fakeServer.Teardown()

	fakeServer.Mux.HandleFunc("/mycontainer", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "HEAD")

		w.Header().Add("X-Container-Read", ".r:*")
		w.WriteHeader(http.StatusNoContent)
	})

	d := &openstackCollector{
		clients: clients{
			storageClient: client.ServiceClient(fakeServer),
		},
	}

	got, err := d.handleObjectStorage(&containers.Container{Name: "mycontainer"})

	assert.NoError(t, err)

	want := &ontology.ObjectStorage{
		Id:           new("mycontainer"),
		Name:         new("mycontainer"),
		ParentId:     new(client.ServiceClient(fakeServer).Endpoint),
		PublicAccess: new(true),
	}
	assert.Equal(t, want, got.(*ontology.ObjectStorage))
}

func Test_openstackCollector_handleObjectStorageService(t *testing.T) {
	fakeServer := testhelper.SetupHTTP()
	defer fakeServer.Teardown()

	d := &openstackCollector{
		clients: clients{
			storageClient: client.ServiceClient(fakeServer),
		},
		region:  "test region",
		project: &project{projectID: "test-project-id"},
	}

	got, err := d.handleObjectStorageService()

	assert.NoError(t, err)

	want := &ontology.ObjectStorageService{
		Id:   new(client.ServiceClient(fakeServer).Endpoint),
		Name: new("Swift Object Storage Service"),
		GeoLocation: &ontology.GeoLocation{
			Region: new("test region"),
		},
		ParentId: new("test-project-id"),
		HttpEndpoint: &ontology.HttpEndpoint{
			Url: new(client.ServiceClient(fakeServer).Endpoint),
		},
	}
	assert.Equal(t, want, got.(*ontology.ObjectStorageService))
}
