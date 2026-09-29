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
	"context"
	"log/slog"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/core/api/ontology"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleNetworkInterfaces creates a network interface resource based on the CSC Hub Ontology
func (d *openstackCollector) handleNetworkInterfaces(network *networks.Network) (ontology.IsResource, error) {
	var (
		l3FirewallEnabled   bool
		restrictedPortsList = []string{}
	)

	// Check if any port associated with the network has security groups enabled. If at least one port has
	// security groups, we consider the L3 firewall to be enabled for the entire network. Note that Neutron
	// security groups only express allow rules (with an implicit deny for everything else), so we cannot derive
	// a meaningful set of restricted ports from them. As a workaround, we will check the security group rules for each port and see if any of the restricted ports (22, 80, 443) are not allowed by any ingress rule. If at least one of these ports is not allowed, we will consider it to be restricted.
	err := ports.List(d.clients.networkClient, ports.ListOpts{
		NetworkID: network.ID,
	}).EachPage(context.Background(), func(_ context.Context, page pagination.Page) (bool, error) {
		portList, err := ports.ExtractPorts(page)
		if err != nil {
			return false, err
		}

		for _, port := range portList {
			if len(port.SecurityGroups) > 0 {
				l3FirewallEnabled = true
				restrictedPorts, err := d.getRestrictedPorts(port.SecurityGroups)
				if err != nil {
					slog.Error("error getting restricted ports for port", slog.String("id", port.ID), tint.Err(err))
				}
				restrictedPortsList = append(restrictedPortsList, restrictedPorts...)
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		log.Error("error listing ports for network", slog.String("id", network.ID), tint.Err(err))
	}

	r := &ontology.NetworkInterface{
		Id:           new(network.ID),
		Name:         new(network.Name),
		Description:  new(network.Description),
		CreationTime: timestamppb.New(network.CreatedAt),
		GeoLocation: &ontology.GeoLocation{
			Region: new(d.region),
		},
		Labels:   labels(new(network.Tags)),
		ParentId: new(network.ProjectID),
		Raw:      new(collector.Raw(network)),
		AccessRestriction: &ontology.AccessRestriction{
			Type: &ontology.AccessRestriction_L3Firewall{
				L3Firewall: &ontology.L3Firewall{
					Enabled:         new(l3FirewallEnabled),
					RestrictedPorts: restrictedPortsList,
				},
			},
		},
	}

	log.Info("Adding network interface", slog.String("name", network.Name))

	return r, nil
}
