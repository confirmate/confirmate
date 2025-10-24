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
	"confirmate.io/collectors/cloud/api/ontology"

	"github.com/gocsaf/csaf/v3/csaf"
)

func (d *csafDiscovery) providerTransportEncryption(url string) *ontology.TransportEncryption {
	res, err := d.client.Get(url)
	if err != nil {
		return &ontology.TransportEncryption{
			Enabled: false,
		}
	}

	return transportEncryption(res.TLS)
}

func providerValidationErrors(messages csaf.ProviderMetadataLoadMessages) (errs []*ontology.Error) {
	for _, m := range messages {
		errs = append(errs, &ontology.Error{Message: m.Message})
	}
	return
}
