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
	"net/http"
	"testing"

	"confirmate.io/collectors/cloud/api/ontology"
	"confirmate.io/collectors/cloud/internal/constants"
	"confirmate.io/collectors/cloud/internal/testutil/assert"

	"github.com/gocsaf/csaf/v3/csaf"
)

func Test_csafDiscovery_providerTransportEncryption(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	type args struct {
		url string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ontology.TransportEncryption
	}{
		{
			name: "happy path",
			args: args{url: goodProvider.URL},
			fields: fields{
				client: goodProvider.Client(),
			},
			want: &ontology.TransportEncryption{
				Enabled:         true,
				Protocol:        constants.TLS,
				ProtocolVersion: 1.3,
				CipherSuites: []*ontology.CipherSuite{
					{
						MacAlgorithm:  constants.SHA_256,
						SessionCipher: constants.AES_128_GCM,
					},
				},
			},
		},
		{
			name: "fail - bad certificate",
			args: args{url: goodProvider.URL},
			fields: fields{
				client: http.DefaultClient,
			},
			want: &ontology.TransportEncryption{
				Enabled: false,
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
			got := d.providerTransportEncryption(tt.args.url)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_providerValidationErrors(t *testing.T) {
	type args struct {
		messages csaf.ProviderMetadataLoadMessages
	}
	tests := []struct {
		name string
		args args
		want assert.Want[[]*ontology.Error]
	}{
		{
			name: "messages given",
			args: args{
				messages: csaf.ProviderMetadataLoadMessages{
					csaf.ProviderMetadataLoadMessage{
						Message: "message1",
					},
					csaf.ProviderMetadataLoadMessage{
						Message: "message2",
					},
				},
			},
			want: func(t *testing.T, got []*ontology.Error) bool {
				want := []*ontology.Error{
					{
						Message: "message1",
					},
					{
						Message: "message2",
					},
				}
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "no messages given",
			args: args{},
			want: assert.Nil[[]*ontology.Error],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErrs := providerValidationErrors(tt.args.messages)

			tt.want(t, gotErrs)
		})
	}
}
