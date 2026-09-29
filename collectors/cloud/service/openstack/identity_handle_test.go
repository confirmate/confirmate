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
	"time"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/users"
)

func Test_openstackCollector_handleIdentity(t *testing.T) {
	d := &openstackCollector{
		region: "test region",
	}

	tests := []struct {
		name     string
		identity *users.User
		want     *ontology.Identity
	}{
		{
			name: "no password expiration configured",
			identity: &users.User{
				ID:               "test-user-id",
				Name:             "test-user",
				DefaultProjectID: "test-project-id",
				Enabled:          true,
			},
			want: &ontology.Identity{
				Id:   new("test-user-id"),
				Name: new("test-user"),
				GeoLocation: &ontology.GeoLocation{
					Region: new("test region"),
				},
				ParentId:              new("test-project-id"),
				Activated:             new(true),
				Raw:                   new(""),
				DisablePasswordPolicy: new(true),
			},
		},
		{
			name: "password expiration configured",
			identity: &users.User{
				ID:                "test-user-id",
				Name:              "test-user",
				DefaultProjectID:  "test-project-id",
				Enabled:           true,
				PasswordExpiresAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: &ontology.Identity{
				Id:   new("test-user-id"),
				Name: new("test-user"),
				GeoLocation: &ontology.GeoLocation{
					Region: new("test region"),
				},
				ParentId:              new("test-project-id"),
				Activated:             new(true),
				Raw:                   new(""),
				DisablePasswordPolicy: new(false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.handleIdentity(tt.identity)

			assert.NoError(t, err)

			gotNew := got.(*ontology.Identity)
			assert.NotEmpty(t, gotNew.GetRaw())
			gotNew.Raw = new("")

			assert.Equal(t, tt.want, gotNew)
		})
	}
}
