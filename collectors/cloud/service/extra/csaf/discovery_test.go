package csaf

import (
	"net/http"
	"os"
	"testing"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/internal/collectortest/csaf/providertest"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"github.com/gocsaf/csaf/v3/csaf"
)

// validAdvisory contains the structure of a valid CSAF Advisory that validates against the JSON schema
var validAdvisory = &csaf.Advisory{
	Document: &csaf.Document{
		Category:    util.Ref(csaf.DocumentCategory("csaf_security_advisory")),
		CSAFVersion: util.Ref(csaf.CSAFVersion20),
		Title:       util.Ref("Buffer overflow in Test Product"),
		Publisher: &csaf.DocumentPublisher{
			Name:      util.Ref("Test Vendor"),
			Category:  util.Ref(csaf.CSAFCategoryVendor),
			Namespace: util.Ref("http://localhost"),
		},
		Tracking: &csaf.Tracking{
			ID:                 util.Ref(csaf.TrackingID("some-id")),
			CurrentReleaseDate: util.Ref("2020-07-01T10:09:07Z"),
			InitialReleaseDate: util.Ref("2020-07-01T10:09:07Z"),
			Generator: &csaf.Generator{
				Date: util.Ref("2020-07-01T10:09:07Z"),
				Engine: &csaf.Engine{
					Name:    util.Ref("test"),
					Version: util.Ref("1.0"),
				},
			},
			Status:  util.Ref(csaf.CSAFTrackingStatusFinal),
			Version: util.Ref(csaf.RevisionNumber("1")),
			RevisionHistory: csaf.Revisions{
				&csaf.Revision{
					Date:    util.Ref("2020-07-01T10:09:07Z"),
					Number:  util.Ref(csaf.RevisionNumber("1")),
					Summary: util.Ref("First and final version"),
				},
			},
		},
	},
	ProductTree: &csaf.ProductTree{
		Branches: csaf.Branches{
			&csaf.Branch{
				Category: util.Ref(csaf.CSAFBranchCategoryVendor),
				Name:     util.Ref("Test Vendor"),
				Product: &csaf.FullProductName{
					Name:      util.Ref("Test Product"),
					ProductID: util.Ref(csaf.ProductID("CSAFPID-0001")),
				},
			},
		},
	},
}

var goodProvider *providertest.TrustedProvider

func TestMain(m *testing.M) {
	var advisories = map[csaf.TLPLabel][]*csaf.Advisory{
		csaf.TLPLabelWhite: {
			validAdvisory,
		},
	}

	goodProvider = providertest.NewTrustedProvider(
		advisories,
		providertest.NewGoodIndexTxtWriter(),
		func(pmd *csaf.ProviderMetadata) {
			pmd.Publisher = &csaf.Publisher{
				Name:      util.Ref("Test Vendor"),
				Category:  util.Ref(csaf.CSAFCategoryVendor),
				Namespace: util.Ref("http://localhost"),
			}
		})
	defer goodProvider.Close()

	code := m.Run()
	os.Exit(code)
}

func TestNewTrustedProviderCollector(t *testing.T) {
	type args struct {
		opts []CollectorOption
	}
	tests := []struct {
		name string
		args args
		want cloud.Collector
	}{
		{
			name: "Happy path",
			args: args{},
			want: &csafCollector{
				ctID:   config.DefaultTargetOfEvaluationID,
				domain: "confirmate.io",
				client: http.DefaultClient,
			},
		},
		{
			name: "Happy path: with target of evaluation id",
			args: args{
				opts: []CollectorOption{WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1)},
			},
			want: &csafCollector{
				ctID:   testdata.MockTargetOfEvaluationID1,
				domain: "confirmate.io",
				client: http.DefaultClient,
			},
		},
		{
			name: "Happy path: with domain",
			args: args{
				opts: []CollectorOption{WithProviderDomain("mock")},
			},
			want: &csafCollector{
				ctID:   config.DefaultTargetOfEvaluationID,
				client: http.DefaultClient,
				domain: "mock",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTrustedProviderCollector(tt.args.opts...)
			assert.Equal(t, tt.want, got, assert.CompareAllUnexported())
		})
	}
}

func Test_csafCollector_List(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	tests := []struct {
		name     string
		fields   fields
		wantList assert.Want[[]ontology.IsResource]
		wantErr  assert.WantErr
	}{
		{
			name: "fail",
			fields: fields{
				domain: "localhost:1234",
				client: http.DefaultClient,
				ctID:   config.DefaultTargetOfEvaluationID,
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not load provider-metadata.json")
			},
			wantList: assert.Empty[[]ontology.IsResource],
		},
		{
			name: "happy path",
			fields: fields{
				domain: goodProvider.Domain(),
				client: goodProvider.Client(),
				ctID:   config.DefaultTargetOfEvaluationID,
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.NoError(t, err)
			},
			wantList: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				return assert.NotEmpty(t, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{
				domain: tt.fields.domain,
				client: tt.fields.client,
				ctID:   tt.fields.ctID,
			}
			gotList, err := d.List()
			tt.wantErr(t, err)
			tt.wantList(t, gotList)
		})
	}
}

func Test_csafCollector_TargetOfEvaluationID(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Happy path",
			fields: fields{
				ctID: testdata.MockTargetOfEvaluationID1,
			},
			want: testdata.MockTargetOfEvaluationID1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{
				domain: tt.fields.domain,
				ctID:   tt.fields.ctID,
				client: tt.fields.client,
			}
			if got := d.TargetOfEvaluationID(); got != tt.want {
				t.Errorf("csafCollector.TargetOfEvaluationID() = %v, want %v", got, tt.want)
			}
		})
	}
}
