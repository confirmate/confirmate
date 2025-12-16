package openstack

import (
	"testing"

	"confirmate.io/collectors/cloud/internal/collectortest/openstacktest"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
)

func Test_labels(t *testing.T) {
	type args struct {
		tags *[]string
	}
	tests := []struct {
		name string
		args args
		want assert.Want[map[string]string]
	}{
		{
			name: "empty input",
			args: args{},
			want: func(t *testing.T, got map[string]string, msgAndArgs ...any) bool {
				want := map[string]string{}

				return assert.Equal(t, want, got)
			},
		},
		{
			name: "Happy path",
			args: args{
				tags: &[]string{
					"tag1",
					"tag2",
					"tag3",
				},
			},
			want: func(t *testing.T, got map[string]string, msgAndArgs ...any) bool {
				want := map[string]string{
					"tag1": "",
					"tag2": "",
					"tag3": "",
				}

				return assert.Equal(t, want, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labels(tt.args.tags)

			tt.want(t, got)
		})
	}
}

func Test_openstackCollector_getAttachedNetworkInterfaces(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	type fields struct {
		ctID       string
		clients    clients
		authOpts   *gophercloud.AuthOptions
		testhelper bool
	}
	type args struct {
		serverID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[[]string]
		wantErr assert.WantErr
	}{
		{
			name: "error getting network interfaces",
			fields: fields{
				testhelper: false,
				clients: clients{
					provider: &gophercloud.ProviderClient{
						TokenID: client.TokenID,
						EndpointLocator: func(eo gophercloud.EndpointOpts) (string, error) {
							return testhelper.Endpoint(), nil
						},
					},
				},
			},
			args: args{
				serverID: "ef079b0c-e610-4dfb-b1aa-b49f07ac48e5",
			},
			want: assert.Nil[[]string],
			wantErr: func(tt *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not list network interfaces:")
			},
		},
		{
			name: "Happy path",
			fields: fields{
				testhelper: true,
				clients: clients{
					provider: &gophercloud.ProviderClient{
						TokenID: client.TokenID,
						EndpointLocator: func(eo gophercloud.EndpointOpts) (string, error) {
							return testhelper.Endpoint(), nil
						},
					},
				},
			},
			args: args{
				serverID: "ef079b0c-e610-4dfb-b1aa-b49f07ac48e5",
			},
			want: func(t *testing.T, got []string, msgAndArgs ...any) bool {
				return assert.Equal(t, "8a5fe506-7e9f-4091-899b-96336909d93c", got[0])
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
			}

			if tt.fields.testhelper {
				openstacktest.HandleInterfaceListSuccessfully(t)

			}

			got, err := d.getAttachedNetworkInterfaces(tt.args.serverID)

			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}
