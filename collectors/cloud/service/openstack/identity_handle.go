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
	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/core/api/ontology"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/users"
)

// handleIdentity creates an identity resource based on the CSC Hub Ontology
func (d *openstackCollector) handleIdentity(identity *users.User) (ontology.IsResource, error) {
	r := &ontology.Identity{
		Id:   new(identity.ID),
		Name: new(identity.Name),
		GeoLocation: &ontology.GeoLocation{
			Region: new(d.region),
		},
		ParentId:  new(identity.DefaultProjectID),
		Raw:       new(collector.Raw(identity)),
		Activated: new(identity.Enabled),
		// A zero PasswordExpiresAt means Keystone has no password expiration policy configured for this user.
		DisablePasswordPolicy: new(identity.PasswordExpiresAt.IsZero()),
	}

	return r, nil
}
