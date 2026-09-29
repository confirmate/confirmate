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
	"fmt"
	"net/http"
	"testing"
	"time"

	"confirmate.io/collectors/cloud/internal/collectortest/openstacktest"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_openstackCollector_collectBlockStorage(t *testing.T) {
	fakeServer := testhelper.SetupHTTP()
	defer fakeServer.Teardown()

	openstacktest.MockStorageListResponse(t, fakeServer)

	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
		region   string
		domain   *domain
		project  *project
	}
	tests := []struct {
		name     string
		fields   fields
		wantList assert.Want[[]ontology.IsResource]
		wantErr  assert.WantErr
	}{
		{
			name: "Happy path",
			fields: fields{
				authOpts: &gophercloud.AuthOptions{
					IdentityEndpoint: testdata.MockOpenstackIdentityEndpoint,
					Username:         testdata.MockOpenstackUsername,
					Password:         testdata.MockOpenstackPassword,
					TenantName:       testdata.MockOpenstackTenantName,
				},
				clients: clients{
					provider: &gophercloud.ProviderClient{
						TokenID: client.TokenID,
						EndpointLocator: func(eo gophercloud.EndpointOpts) (string, error) {
							return fakeServer.Endpoint(), nil
						},
					},
					blockStorageClient: client.ServiceClient(fakeServer),
				},
				region:  "test region",
				domain:  &domain{},
				project: &project{},
			},
			wantList: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				assert.Equal(t, 2, len(got))

				t1, err := time.Parse("2006-01-02T15:04:05.000000", "2015-09-17T03:35:03.000000")
				assert.NoError(t, err)

				want := &ontology.BlockStorage{
					Id:           new("289da7f8-6440-407c-9fb4-7db01ec49164"),
					Name:         new("vol-001"),
					Description:  new(""),
					CreationTime: timestamppb.New(t1),
					GeoLocation: &ontology.GeoLocation{
						Region: new("test region"),
					},
					ParentId: new("83ec2e3b-4321-422b-8706-a84185f52a0a"),
					Labels:   map[string]string{},
					Raw:      new(""),
				}

				got0 := got[0].(*ontology.BlockStorage)

				assert.NotEmpty(t, got0.GetRaw())
				got0.Raw = new("")
				return assert.Equal(t, want, got0)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &openstackCollector{
				ctID:     tt.fields.ctID,
				clients:  tt.fields.clients,
				authOpts: tt.fields.authOpts,
				region:   tt.fields.region,
				domain:   tt.fields.domain,
				project:  tt.fields.project,
			}
			gotList, err := d.collectBlockStorage()

			tt.wantList(t, gotList)
			tt.wantErr(t, err)
		})
	}
}

func Test_openstackCollector_collectObjectStorage(t *testing.T) {
	fakeServer := testhelper.SetupHTTP()
	defer fakeServer.Teardown()

	fakeServer.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "GET")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.URL.Query().Get("marker") == "" {
			fmt.Fprint(w, `[{"name": "mycontainer", "count": 0, "bytes": 0}]`)
		} else {
			fmt.Fprint(w, `[]`)
		}
	})
	fakeServer.Mux.HandleFunc("/mycontainer", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "HEAD")

		w.Header().Add("X-Container-Read", ".r:*")
		w.WriteHeader(http.StatusNoContent)
	})

	d := &openstackCollector{
		clients: clients{
			storageClient: client.ServiceClient(fakeServer),
		},
		project: &project{},
	}

	gotList, err := d.collectObjectStorage()

	assert.NoError(t, err)
	assert.Equal(t, 1, len(gotList))

	got0 := gotList[0].(*ontology.ObjectStorage)
	want := &ontology.ObjectStorage{
		Id:           new("mycontainer"),
		Name:         new("mycontainer"),
		ParentId:     new(client.ServiceClient(fakeServer).Endpoint),
		PublicAccess: new(true),
	}
	assert.Equal(t, want, got0)
}

func Test_openstackCollector_collectObjectStorageService(t *testing.T) {
	fakeServer := testhelper.SetupHTTP()
	defer fakeServer.Teardown()

	d := &openstackCollector{
		clients: clients{
			storageClient: client.ServiceClient(fakeServer),
		},
		region:  "test region",
		project: &project{projectID: "test-project-id"},
	}

	gotList, err := d.collectObjectStorageService()

	assert.NoError(t, err)
	assert.Equal(t, 1, len(gotList))

	got0 := gotList[0].(*ontology.ObjectStorageService)
	assert.Equal(t, "Swift Object Storage Service", *got0.Name)
	assert.Equal(t, "test region", *got0.GeoLocation.Region)
	assert.Equal(t, new("test-project-id"), got0.ParentId)
}
