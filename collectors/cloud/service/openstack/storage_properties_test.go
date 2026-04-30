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

package openstack

import (
	"testing"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
)

func Test_getParentID(t *testing.T) {
	type args struct {
		volume *volumes.Volume
	}
	tests := []struct {
		name string
		args args
		want assert.Want[string]
	}{
		{
			name: "Happy path: no attached server available",
			args: args{
				&volumes.Volume{
					TenantID: testdata.MockOpenstackVolumeTenantID,
				},
			},
			want: func(t *testing.T, got string, msgAndArgs ...any) bool {
				return assert.Equal(t, testdata.MockOpenstackVolumeTenantID, got)
			},
		},
		{
			name: "Happy path: attached serverID",
			args: args{
				&volumes.Volume{
					TenantID: testdata.MockOpenstackVolumeTenantID,
					Attachments: []volumes.Attachment{
						{
							ServerID: testdata.MockOpenstackServerID1,
						},
					},
				},
			},
			want: func(t *testing.T, got string, msgAndArgs ...any) bool {
				return assert.Equal(t, testdata.MockOpenstackServerID1, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getParentID(tt.args.volume)

			tt.want(t, got)
		})
	}
}
