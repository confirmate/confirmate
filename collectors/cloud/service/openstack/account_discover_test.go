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
	"testing"

	"confirmate.io/collectors/cloud/internal/collectortest/openstacktest"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
)

func Test_openstackCollector_collectProjects(t *testing.T) {
	fakeServer := testhelper.SetupHTTP()
	defer fakeServer.Teardown()
	openstacktest.HandleListProjectsSuccessfully(t, fakeServer)

	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
		region   string
		domain   *domain
		project  *project
	}
	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[[]ontology.IsResource]
		wantErr assert.WantErr
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
					identityClient: client.ServiceClient(fakeServer),
				},
				region:  "test region",
				domain:  &domain{},
				project: &project{},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				assert.Equal(t, 2, len(got))

				want := &ontology.ResourceGroup{
					Id:          new("1234"),
					Name:        new("Red Team"),
					Description: new("The team that is red"),
					GeoLocation: &ontology.GeoLocation{
						Region: new("test region"),
					},
					Labels: map[string]string{
						"Red":  "",
						"Team": "",
					},
					ParentId: new(""),
					Raw:      new(""),
				}

				got0 := got[0].(*ontology.ResourceGroup)

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

			gotList, err := d.collectProjects()

			tt.want(t, gotList)
			tt.wantErr(t, err)
		})
	}
}

func Test_openstackCollector_collectDomain(t *testing.T) {
	fakeServer := testhelper.SetupHTTP()
	defer fakeServer.Teardown()
	openstacktest.HandleListDomainsSuccessfully(t, fakeServer)

	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
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
					identityClient: client.ServiceClient(fakeServer),
				},
				domain:  &domain{},
				project: &project{},
			},
			wantList: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				assert.Equal(t, 2, len(got))

				want := &ontology.Account{
					Id:          new("2844b2a08be147a08ef58317d6471f1f"),
					Name:        new("domain one"),
					Description: new("some description"),
					Raw:         new(""),
				}

				got0 := got[0].(*ontology.Account)

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
				domain:   tt.fields.domain,
				project:  tt.fields.project,
			}
			gotList, err := d.collectDomains()

			tt.wantList(t, gotList)
			tt.wantErr(t, err)
		})
	}
}
