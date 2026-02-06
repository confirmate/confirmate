package openstack

import (
	"testing"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
)

func Test_openstackCollector_handleProject(t *testing.T) {
	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
		region   string
		domain   *domain
		project  *project
	}
	type args struct {
		project *projects.Project
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
				region: "test region",
			},
			args: args{
				project: &projects.Project{
					ID:          testdata.MockOpenstackProjectID1,
					Name:        testdata.MockOpenstackProjectName1,
					Description: testdata.MockOpenstackProjectDescription1,
					Tags:        []string{},
					ParentID:    testdata.MockOpenstackProjectParentID1,
				},
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				want := &ontology.ResourceGroup{
					Id:   testdata.MockOpenstackProjectID1,
					Name: testdata.MockOpenstackProjectName1,
					GeoLocation: &ontology.GeoLocation{
						Region: "test region",
					},
					Description: testdata.MockOpenstackProjectDescription1,
					Labels:      labels(util.Ref([]string{})),
					ParentId:    util.Ref(testdata.MockOpenstackProjectParentID1),
				}

				gotNew := got.(*ontology.ResourceGroup)
				assert.NotEmpty(t, gotNew.GetRaw())
				gotNew.Raw = ""
				return assert.Equal(t, want, gotNew)
			},
			wantErr: assert.NoError,
		}}
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
			got, err := d.handleProject(tt.args.project)

			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}

func Test_openstackCollector_handleDomain(t *testing.T) {
	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
		domain   *domain
		project  *project
	}
	type args struct {
		domain *domains.Domain
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
			args: args{
				domain: &domains.Domain{
					ID:          testdata.MockOpenstackDomainID1,
					Name:        testdata.MockOpenstackDomainName1,
					Description: testdata.MockOpenstackDomainDescription1,
				},
			},
			want: func(t *testing.T, got ontology.IsResource, msgAndArgs ...any) bool {
				want := &ontology.Account{
					Id:          testdata.MockOpenstackDomainID1,
					Name:        testdata.MockOpenstackDomainName1,
					Description: testdata.MockOpenstackDomainDescription1,
				}

				gotNew := got.(*ontology.Account)
				assert.NotEmpty(t, gotNew.GetRaw())
				gotNew.Raw = ""
				return assert.Equal(t, want, gotNew)
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
			got, err := d.handleDomain(tt.args.domain)

			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}
