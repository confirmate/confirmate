package openstack

import (
	"testing"
	"time"

	"confirmate.io/collectors/cloud/api/ontology"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/collectors/cloud/internal/testutil/assert"
	"confirmate.io/collectors/cloud/internal/testutil/servicetest/discoverytest/openstacktest"
	"confirmate.io/collectors/cloud/internal/util"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_openstackDiscovery_discoverNetworkInterfaces(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()
	openstacktest.HandleNetworkListSuccessfully(t)

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
		wantErr  assert.ErrorAssertionFunc
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
							return testhelper.Endpoint(), nil
						},
					},
					networkClient: client.ServiceClient(),
				},
				region:  "test region",
				domain:  &domain{},
				project: &project{},
			},
			wantList: func(t *testing.T, got []ontology.IsResource) bool {
				assert.Equal(t, 2, len(got))

				t1, err := time.Parse("2006-01-02T15:04:05", "2019-06-30T04:15:37")
				assert.NoError(t, err)

				want := &ontology.NetworkInterface{
					Id:           "d32019d3-bc6e-4319-9c1d-6722fc136a22",
					Name:         "public",
					CreationTime: timestamppb.New(t1),
					GeoLocation: &ontology.GeoLocation{
						Region: "test region",
					},
					Labels:   map[string]string{},
					ParentId: util.Ref("4fd44f30292945e481c7b8a0c8908869"),
				}

				got0 := got[0].(*ontology.NetworkInterface)

				assert.NotEmpty(t, got0.GetRaw())
				got0.Raw = ""
				return assert.Equal(t, want, got0)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &openstackDiscovery{
				ctID:     tt.fields.ctID,
				clients:  tt.fields.clients,
				authOpts: tt.fields.authOpts,
				region:   tt.fields.region,
				domain:   tt.fields.domain,
				project:  tt.fields.project,
			}
			gotList, err := d.discoverNetworkInterfaces()

			tt.wantList(t, gotList)
			tt.wantErr(t, err)
		})
	}
}
