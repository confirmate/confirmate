// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package csaf

import (
	"reflect"
	"testing"

	"confirmate.io/core/api/ontology"
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
