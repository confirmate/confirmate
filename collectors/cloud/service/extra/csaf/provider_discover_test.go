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
	"reflect"
	"testing"

	"confirmate.io/collectors/cloud/api/ontology"
)

func Test_getIDsOf(t *testing.T) {
	type args struct {
		documents []ontology.IsResource
	}
	tests := []struct {
		name    string
		args    args
		wantIds []string
	}{
		{
			name:    "no documents given",
			args:    args{},
			wantIds: nil,
		},
		{
			name: "documents given",
			args: args{
				documents: []ontology.IsResource{
					&ontology.SecurityAdvisoryDocument{
						Id: "https://xx.yy.zz/XXX",
					},
				},
			},
			wantIds: []string{"https://xx.yy.zz/XXX"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIds := getIDsOf(tt.args.documents); !reflect.DeepEqual(gotIds, tt.wantIds) {
				t.Errorf("getIDsOf() = %v, want %v", gotIds, tt.wantIds)
			}
		})
	}
}
