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
	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleNetworkInterface creates a network interface resource based on the CSC Hub Ontology
func (d *ionosCollector) handleNetworkInterface(nic ionoscloud.Nic, dc ionoscloud.Datacenter) (ontology.IsResource, error) {
	r := &ontology.NetworkInterface{
		Id:           nic.Id,
		Name:         nic.Properties.Name,
		CreationTime: timestamppb.New(pointer.Deref(nic.Metadata.GetCreatedDate())),
		GeoLocation:  &ontology.GeoLocation{Region: dc.Properties.GetLocation()},
		ParentId:     dc.GetId(),
		Raw:          new(collector.Raw(nic, dc)),
		AccessRestriction: &ontology.AccessRestriction{
			Type: &ontology.AccessRestriction_L3Firewall{
				L3Firewall: &ontology.L3Firewall{
					Enabled: nic.Properties.GetFirewallActive(),
				},
			},
		},
	}

	return r, nil
}
