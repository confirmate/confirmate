// Copyright 2025 Fraunhofer AISEC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package csaf

import (
	"fmt"
	"net/http"
	"testing"

	"confirmate.io/collectors/cloud/api/ontology"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/crypto/openpgp"
	"confirmate.io/collectors/cloud/internal/testutil/assert"
	"github.com/gocsaf/csaf/v3/csaf"
)

func Test_csafDiscovery_handleAdvisory(t *testing.T) {
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
			wantDoc: func(t *testing.T, got *ontology.SecurityAdvisoryDocument) bool {
				// Some debugging output, that can easily be used in Rego
				fmt.Println(ontology.ToPrettyJSON(got))
				return assert.Equal(t, "some-id", got.Id)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafDiscovery{
				domain: tt.fields.domain,
				ctID:   tt.fields.ctID,
				client: tt.fields.client,
			}
			gotDoc, err := d.handleAdvisory(tt.args.label, tt.args.file, tt.args.keyring, tt.args.parentId)
			if (err != nil) != tt.wantErr {
				t.Errorf("csafDiscovery.handleAdvisory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantDoc(t, gotDoc)
		})
	}
}

func Test_csafDiscovery_discoverSecurityAdvisories(t *testing.T) {
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
			wantDocuments: func(t *testing.T, got []ontology.IsResource) bool {
				return assert.NotEmpty(t, got) && assert.Equal(t, "some-id", got[0].GetId())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafDiscovery{
				domain: tt.fields.domain,
				ctID:   tt.fields.ctID,
				client: tt.fields.client,
			}
			gotDocuments, err := d.discoverSecurityAdvisories(tt.args.md, tt.args.keyring, tt.args.parentId)
			if (err != nil) != tt.wantErr {
				t.Errorf("csafDiscovery.discoverSecurityAdvisories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantDocuments(t, gotDocuments)
		})
	}
}
