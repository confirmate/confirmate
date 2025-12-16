package csaf

import (
	"fmt"
	"net/http"
	"testing"

	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/crypto/openpgp"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"
	"github.com/gocsaf/csaf/v3/csaf"
)

func Test_csafCollector_handleAdvisory(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	type args struct {
		label    csaf.TLPLabel
		file     csaf.AdvisoryFile
		keyring  openpgp.EntityList
		parentId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantDoc assert.Want[*ontology.SecurityAdvisoryDocument]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				domain: goodProvider.Domain(),
				ctID:   config.DefaultTargetOfEvaluationID,
				client: goodProvider.Client(),
			},
			args: args{
				label: csaf.TLPLabelWhite,
				file: csaf.DirectoryAdvisoryFile{
					Path: goodProvider.URL + "/.well-known/csaf/white/2020/some-id.json",
				},
				keyring: goodProvider.Keyring,
			},
			wantDoc: func(t *testing.T, got *ontology.SecurityAdvisoryDocument, msgAndargs ...any) bool {
				// Some debugging output, that can easily be used in Rego
				fmt.Println(ontology.ToPrettyJSON(got))
				return assert.Equal(t, "some-id", got.Id)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{
				domain: tt.fields.domain,
				ctID:   tt.fields.ctID,
				client: tt.fields.client,
			}
			gotDoc, err := d.handleAdvisory(tt.args.label, tt.args.file, tt.args.keyring, tt.args.parentId)
			if (err != nil) != tt.wantErr {
				t.Errorf("csafCollector.handleAdvisory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantDoc(t, gotDoc)
		})
	}
}

func Test_csafCollector_collectSecurityAdvisories(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	type args struct {
		md       *csaf.LoadedProviderMetadata
		keyring  openpgp.EntityList
		parentId string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantDocuments assert.Want[[]ontology.IsResource]
		wantErr       bool
	}{
		{
			name: "happy path",
			fields: fields{
				domain: goodProvider.Domain(),
				ctID:   config.DefaultTargetOfEvaluationID,
				client: goodProvider.Client(),
			},
			args: args{
				md: &csaf.LoadedProviderMetadata{
					URL:      goodProvider.WellKnownProviderURL(),
					Document: goodProvider.DocumentAny(),
				},
			},
			wantDocuments: func(t *testing.T, got []ontology.IsResource, msgAndargs ...any) bool {
				return assert.NotEmpty(t, got) && assert.Equal(t, "some-id", got[0].GetId())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{
				domain: tt.fields.domain,
				ctID:   tt.fields.ctID,
				client: tt.fields.client,
			}
			gotDocuments, err := d.collectSecurityAdvisories(tt.args.md, tt.args.keyring, tt.args.parentId)
			if (err != nil) != tt.wantErr {
				t.Errorf("csafCollector.collectSecurityAdvisories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantDocuments(t, gotDocuments)
		})
	}
}
