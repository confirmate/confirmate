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

package azure

import (
	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func (d *azureCollector) handleLoadBalancer(lb *armnetwork.LoadBalancer) ontology.IsResource {
	return &ontology.LoadBalancer{
		Id:           resourceID(lb.ID),
		Name:         pointer.Deref(lb.Name),
		CreationTime: nil, // No creation time available
		GeoLocation: &ontology.GeoLocation{
			Region: pointer.Deref(lb.Location),
		},
		Labels:   labels(lb.Tags),
		ParentId: resourceGroupID(lb.ID),
		Raw:      collector.Raw(lb),
		Ips:      publicIPAddressFromLoadBalancer(lb),
		Ports:    loadBalancerPorts(lb), // TODO(oxisto): ports should be uint16, not 32
	}
}

// handleApplicationGateway returns the application gateway with its properties
// NOTE: handleApplicationGateway uses the LoadBalancer for now until there is a own resource
func (d *azureCollector) handleApplicationGateway(ag *armnetwork.ApplicationGateway) ontology.IsResource {
	firewallStatus := false

	if ag.Properties != nil && ag.Properties.WebApplicationFirewallConfiguration != nil {
		firewallStatus = pointer.Deref(ag.Properties.WebApplicationFirewallConfiguration.Enabled)
	}

	return &ontology.LoadBalancer{
		Id:           resourceID(ag.ID),
		Name:         pointer.Deref(ag.Name),
		CreationTime: nil, // No creation time available
		GeoLocation: &ontology.GeoLocation{
			Region: pointer.Deref(ag.Location),
		},
		Labels:   labels(ag.Tags),
		ParentId: resourceGroupID(ag.ID),
		Raw:      collector.Raw(ag),
		AccessRestriction: &ontology.AccessRestriction{
			Type: &ontology.AccessRestriction_WebApplicationFirewall{
				WebApplicationFirewall: &ontology.WebApplicationFirewall{
					Enabled: firewallStatus,
				},
			},
		},
	}
}

func (d *azureCollector) handleNetworkInterfaces(ni *armnetwork.Interface) ontology.IsResource {
	return &ontology.NetworkInterface{
		Id:           resourceID(ni.ID),
		Name:         pointer.Deref(ni.Name),
		CreationTime: nil, // No creation time available
		GeoLocation: &ontology.GeoLocation{
			Region: pointer.Deref(ni.Location),
		},
		Labels:   labels(ni.Tags),
		ParentId: resourceGroupID(ni.ID),
		Raw:      collector.Raw(ni),
		AccessRestriction: &ontology.AccessRestriction{
			Type: &ontology.AccessRestriction_L3Firewall{
				L3Firewall: &ontology.L3Firewall{
					Enabled: d.nsgFirewallEnabled(ni),
				},
			},
		},
	}
}
