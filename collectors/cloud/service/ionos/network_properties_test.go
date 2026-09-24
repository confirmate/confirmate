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

func Test_ionosCollector_getRestrictedPorts(t *testing.T) {
	d := &ionosCollector{}

	tests := []struct {
		name string
		nic  ionoscloud.Nic
		want []string
	}{
		{
			name: "no firewall rules",
			nic:  ionoscloud.Nic{},
			want: nil,
		},
		{
			name: "all ports restricted",
			nic: ionoscloud.Nic{
				Entities: &ionoscloud.NicEntities{
					Firewallrules: &ionoscloud.FirewallRules{
						Items: &[]ionoscloud.FirewallRule{
							{Properties: &ionoscloud.FirewallruleProperties{}},
						},
					},
				},
			},
			want: []string{"all"},
		},
		{
			name: "single port",
			nic: ionoscloud.Nic{
				Entities: &ionoscloud.NicEntities{
					Firewallrules: &ionoscloud.FirewallRules{
						Items: &[]ionoscloud.FirewallRule{
							{Properties: &ionoscloud.FirewallruleProperties{
								PortRangeStart: new(int32(22)),
								PortRangeEnd:   new(int32(22)),
							}},
						},
					},
				},
			},
			want: []string{"22"},
		},
		{
			name: "port range",
			nic: ionoscloud.Nic{
				Entities: &ionoscloud.NicEntities{
					Firewallrules: &ionoscloud.FirewallRules{
						Items: &[]ionoscloud.FirewallRule{
							{Properties: &ionoscloud.FirewallruleProperties{
								PortRangeStart: new(int32(80)),
								PortRangeEnd:   new(int32(82)),
							}},
						},
					},
				},
			},
			want: []string{"80", "81", "82"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.getRestrictedPorts(tt.nic)
			assert.Equal(t, tt.want, got)
		})
	}
}
