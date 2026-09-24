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

package ionos

import (
	"testing"

	"confirmate.io/core/util/assert"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func Test_labels(t *testing.T) {
	tests := []struct {
		name string
		lr   ionoscloud.LabelResources
		want map[string]string
	}{
		{
			name: "no items",
			lr:   ionoscloud.LabelResources{},
			want: map[string]string{},
		},
		{
			name: "with items",
			lr: ionoscloud.LabelResources{
				Items: &[]ionoscloud.LabelResource{
					{
						Properties: &ionoscloud.LabelResourceProperties{
							Key:   new("env"),
							Value: new("prod"),
						},
					},
				},
			},
			want: map[string]string{"env": "prod"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labels(tt.lr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_hasEmptySegmentInURL(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "no empty segment",
			s:    "/datacenters/99d85e98-c3da-11ed-afa1-0242ac120002/servers",
			want: false,
		},
		{
			name: "empty segment",
			s:    "/datacenters//servers",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasEmptySegmentInURL(tt.s))
		})
	}
}
