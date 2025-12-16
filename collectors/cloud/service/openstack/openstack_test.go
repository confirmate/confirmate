package openstack

import (
	"errors"
	"fmt"
	"testing"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/internal/collectortest/openstacktest"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
)

func TestNewOpenstackCollector(t *testing.T) {
	type args struct {
		opts []CollectorOption
	}
	tests := []struct {
		name string
		args args
		want assert.Want[cloud.Collector]
	}{
		{
			name: "error: oauthOpts not set",
			args: args{},
			want: assert.Nil[cloud.Collector],
		},
		{
			name: "Happy path: Name",
			args: args{
				opts: []CollectorOption{
					WithAuthorizer(gophercloud.AuthOptions{
						IdentityEndpoint: testdata.MockOpenstackIdentityEndpoint,
						Username:         testdata.MockOpenstackUsername,
						Password:         testdata.MockOpenstackPassword,
						TenantName:       testdata.MockOpenstackTenantName,
						AllowReauth:      true,
					}),
					WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID2),
				},
			},
			want: func(t *testing.T, got cloud.Collector, msgAndargs ...any) bool {
				return assert.Equal(t, "OpenStack", got.Name())
			},
		},
		{
			name: "Happy path: with target of evaluation id",
			args: args{
				opts: []CollectorOption{
					WithAuthorizer(gophercloud.AuthOptions{
						IdentityEndpoint: testdata.MockOpenstackIdentityEndpoint,
						Username:         testdata.MockOpenstackUsername,
						Password:         testdata.MockOpenstackPassword,
						TenantName:       testdata.MockOpenstackTenantName,
						AllowReauth:      true,
					}),
					WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID2),
				},
			},
			want: func(t *testing.T, got cloud.Collector, msgAndargs ...any) bool {
				assert.Equal(t, testdata.MockTargetOfEvaluationID2, got.TargetOfEvaluationID())
				return assert.NotNil(t, got)
			},
		},
		{
			name: "Happy path: with authorizer",
			args: args{
				opts: []CollectorOption{
					WithAuthorizer(gophercloud.AuthOptions{
						IdentityEndpoint: testdata.MockOpenstackIdentityEndpoint,
						Username:         testdata.MockOpenstackUsername,
						Password:         testdata.MockOpenstackPassword,
						TenantName:       testdata.MockOpenstackTenantName,
						AllowReauth:      true,
					}),
				},
			},
			want: func(t *testing.T, got cloud.Collector, msgAndargs ...any) bool {
				assert.Equal(t, config.DefaultTargetOfEvaluationID, got.TargetOfEvaluationID())
				return assert.NotNil(t, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOpenstackCollector(tt.args.opts...)
			tt.want(t, got)
		})
	}
}

func Test_openstackCollector_authorize(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.WantErr
	}{
		{
			name:   "error authentication",
			fields: fields{},
			wantErr: func(tt *testing.T, err error, msgAndargs ...any) bool {
				return assert.ErrorContains(t, err, "error while authenticating:")
			},
		},
		{
			name: "compute client error",
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
							return "", errors.New("this is a test error")
						},
					},
				},
			},
			wantErr: func(tt *testing.T, err error, msgAndargs ...any) bool {
				return assert.ErrorContains(t, err, "could not create compute client:")
			},
		},
		{
			name: "network client error",
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
							if eo.Type == "network" {
								return "", errors.New("this is a test error")
							}
							return testhelper.Endpoint(), nil
						},
					},
				},
			},
			wantErr: func(tt *testing.T, err error, msgAndargs ...any) bool {
				return assert.ErrorContains(t, err, "could not create network client:")
			},
		},
		{
			name: "storage client error",
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
							if eo.Type == "block-storage" {
								return "", errors.New("this is a test error")
							}
							return testhelper.Endpoint(), nil
						},
					},
				},
			},
			wantErr: func(tt *testing.T, err error, msgAndargs ...any) bool {
				return assert.ErrorContains(t, err, "could not create block storage client:")
			},
		},
		{
			name: "cluster client error",
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
							if eo.Type == "container-infrastructure-management" {
								return "", errors.New("this is a test error")
							}
							return testhelper.Endpoint(), nil
						},
					},
				},
			},
			wantErr: func(tt *testing.T, err error, msgAndargs ...any) bool {
				return assert.ErrorContains(t, err, "could not create cluster client:")
			},
		},
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
				},
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

			err := d.authorize()

			tt.wantErr(t, err)
		})
	}
}

func TestNewAuthorizer(t *testing.T) {
	type envVariables struct {
		envVariableKey   string
		envVariableValue string
	}
	type fields struct {
		envVariables []envVariables
	}
	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[gophercloud.AuthOptions]
		wantErr assert.WantErr
	}{
		{
			name: "error: missing OS_AUTH_URL",
			fields: fields{
				envVariables: []envVariables{},
			},
			want: func(t *testing.T, got gophercloud.AuthOptions, msgAndargs ...any) bool {
				assert.True(t, got.AllowReauth)
				got.AllowReauth = false // We do not want to check this field in the following
				return assert.Empty(t, got)
			},
			wantErr: func(tt *testing.T, err error, msgAndargs ...any) bool {
				return assert.ErrorContains(t, err, "Missing environment variable [OS_AUTH_URL]")
			},
		},
		{
			name: "Happy path",
			fields: fields{
				envVariables: []envVariables{
					{
						envVariableKey:   "OS_AUTH_URL",
						envVariableValue: testdata.MockOpenstackIdentityEndpoint,
					},
					{
						envVariableKey:   "OS_USERNAME",
						envVariableValue: testdata.MockOpenstackUsername,
					},
					{
						envVariableKey:   "OS_PASSWORD",
						envVariableValue: testdata.MockOpenstackPassword,
					},
					{
						envVariableKey:   "OS_TENANT_ID",
						envVariableValue: testdata.MockOpenstackProjectID1,
					},
					{
						envVariableKey:   "OS_PROJECT_ID",
						envVariableValue: testdata.MockOpenstackProjectID1,
					},
				},
			},
			want: func(t *testing.T, got gophercloud.AuthOptions, msgAndargs ...any) bool {
				want := gophercloud.AuthOptions{
					IdentityEndpoint: testdata.MockOpenstackIdentityEndpoint,
					Username:         testdata.MockOpenstackUsername,
					Password:         testdata.MockOpenstackPassword,
					TenantID:         testdata.MockOpenstackProjectID1,
					AllowReauth:      true,
				}
				return assert.Equal(t, want, got)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env variables
			for _, env := range tt.fields.envVariables {
				if env.envVariableKey != "" {
					t.Setenv(env.envVariableKey, env.envVariableValue)
				}
			}
			got, err := NewAuthorizer()

			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}

func Test_openstackCollector_List(t *testing.T) {
	testhelper.SetupHTTP()

	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
		domain   *domain
		project  *project

		testhelper string
	}
	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[[]ontology.IsResource]
		wantErr assert.WantErr
	}{
		{
			name: "error collect server",
			fields: fields{
				testhelper: "server",
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
					identityClient: client.ServiceClient(),
				},
				project: &project{},
				domain:  &domain{},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndargs ...any) bool {
				return assert.Equal(t, 0, len(got))
			},
			wantErr: assert.NoError,
		},
		{
			name: "error collect network interfaces",
			fields: fields{
				testhelper: "network",
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
					identityClient: client.ServiceClient(),
				},
				project: &project{},
				domain:  &domain{},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				return assert.Equal(t, 4, len(got))
			},
			wantErr: assert.NoError,
		},
		{
			name: "error collect block storage",
			fields: fields{
				testhelper: "storage",
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
					identityClient: client.ServiceClient(),
				},
				project: &project{},
				domain:  &domain{},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				return assert.Equal(t, 6, len(got))
			},
			wantErr: assert.NoError,
		},
		{
			name: "error collect clusters",
			fields: fields{
				testhelper: "clusters",
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
					identityClient: client.ServiceClient(),
				},
				project: &project{},
				domain:  &domain{},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				return assert.Equal(t, 8, len(got))
			},
			wantErr: assert.NoError,
		},
		{
			name: "error collect projects: but there is no error, as a resource is added based on other information collected before.",
			fields: fields{
				testhelper: "project",
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
					identityClient: client.ServiceClient(),
				},
				project: &project{},
				domain: &domain{
					domainID: "test domain ID",
				},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				want := &ontology.ResourceGroup{
					Id:       "fcad67a6189847c4aecfa3c81a05783b",
					Name:     "fcad67a6189847c4aecfa3c81a05783b",
					ParentId: util.Ref("test domain ID"),
					Raw:      "",
				}

				got0 := got[9].(*ontology.ResourceGroup)
				assert.NotEmpty(t, got0.GetRaw())
				got0.Raw = ""
				return assert.Equal(t, want, got0)
			},
			wantErr: assert.NoError,
		},
		{
			name: "error collect domains",
			fields: fields{
				testhelper: "domain",
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
					identityClient: client.ServiceClient(),
				},
				project: &project{},
				domain:  &domain{},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				return assert.Equal(t, 10, len(got))
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path",
			fields: fields{
				testhelper: "all",
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
					identityClient: client.ServiceClient(),
				},
				project: &project{},
				domain: &domain{
					domainID: "test domain ID",
				},
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				return assert.Equal(t, 11, len(got))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.SetupHTTP()

			d := &openstackCollector{
				ctID:     tt.fields.ctID,
				clients:  tt.fields.clients,
				authOpts: tt.fields.authOpts,
				domain:   tt.fields.domain,
				project:  tt.fields.project,
			}

			switch tt.fields.testhelper {
			case "all":
				fmt.Println("Setting up handlers for all resources")
				const ConsoleOutputBody = `{
					"output": "output test"
				}`

				openstacktest.HandleServerListSuccessfully(t)
				openstacktest.HandleShowConsoleOutputSuccessfully(t, ConsoleOutputBody)
				openstacktest.HandleInterfaceListSuccessfully(t)
				openstacktest.HandleNetworkListSuccessfully(t)
				openstacktest.MockStorageListResponse(t)
				openstacktest.HandleListClusterSuccessfully(t)
			case "domain":
				fmt.Println("Setting up handlers to get an error for domain resources")
				const ConsoleOutputBody = `{
					"output": "output test"
				}`

				openstacktest.HandleServerListSuccessfully(t)
				openstacktest.HandleShowConsoleOutputSuccessfully(t, ConsoleOutputBody)
				openstacktest.HandleInterfaceListSuccessfully(t)
				openstacktest.HandleNetworkListSuccessfully(t)
				openstacktest.MockStorageListResponse(t)
				openstacktest.HandleListClusterSuccessfully(t)
			case "project":
				fmt.Println("Setting up handlers to get an error for project resources")
				const ConsoleOutputBody = `{
					"output": "output test"
				}`

				openstacktest.HandleServerListSuccessfully(t)
				openstacktest.HandleShowConsoleOutputSuccessfully(t, ConsoleOutputBody)
				openstacktest.HandleInterfaceListSuccessfully(t)
				openstacktest.HandleNetworkListSuccessfully(t)
				openstacktest.MockStorageListResponse(t)
				openstacktest.HandleListClusterSuccessfully(t)
			case "clusters":
				fmt.Println("Setting up handlers to get an error for storage resources")
				const ConsoleOutputBody = `{
					"output": "output test"
				}`

				openstacktest.HandleServerListSuccessfully(t)
				openstacktest.HandleShowConsoleOutputSuccessfully(t, ConsoleOutputBody)
				openstacktest.HandleInterfaceListSuccessfully(t)
				openstacktest.HandleNetworkListSuccessfully(t)
				openstacktest.MockStorageListResponse(t)
			case "storage":
				fmt.Println("Setting up handlers to get an error for storage resources")
				const ConsoleOutputBody = `{
					"output": "output test"
				}`

				openstacktest.HandleServerListSuccessfully(t)
				openstacktest.HandleShowConsoleOutputSuccessfully(t, ConsoleOutputBody)
				openstacktest.HandleInterfaceListSuccessfully(t)
				openstacktest.HandleNetworkListSuccessfully(t)
			case "network":
				fmt.Println("Setting up handlers to get an error for network resources")
				const ConsoleOutputBody = `{
						"output": "output test"
					}`

				openstacktest.HandleServerListSuccessfully(t)
				openstacktest.HandleShowConsoleOutputSuccessfully(t, ConsoleOutputBody)
				openstacktest.HandleInterfaceListSuccessfully(t)
			}

			gotList, err := d.List()

			tt.want(t, gotList)
			tt.wantErr(t, err)
			testhelper.TeardownHTTP()
		})
	}
}
