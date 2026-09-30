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
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// getRestrictedPorts checks the security group rules for the given security group IDs and returns a list of restricted ports (22, 80, 443) that are not allowed by any ingress rule. It considers only TCP traffic and handles cases where a rule allows all ports or has specific port ranges. It does not return any ports if the security group allows all traffic or if there are no security groups associated with the port. The list of restricted ports is not guaranteed to be exhaustive, as it only checks for the specified ports (22, 80, 443) and does not account for other ports that may be restricted by the security group rules.
func (d *openstackCollector) getRestrictedPorts(portSecurityGroup []string) ([]string, error) {
	var (
		restrictedPortsList []string
		portsToCheck        = []int{22, 80, 443}
		allowedPorts        = make(map[int]bool)
	)

	for _, sgID := range portSecurityGroup {
		pager := groups.List(d.clients.networkClient, groups.ListOpts{
			ID: sgID,
		})

		err := pager.EachPage(context.Background(), func(ctx context.Context, page pagination.Page) (bool, error) {
			sgList, err := groups.ExtractGroups(page)
			if err != nil {
				return false, err
			}

			for _, sg := range sgList {
				for _, rule := range sg.Rules {
					if rule.Direction != "ingress" {
						continue
					}

					// An empty protocol or "any" also applies to TCP traffic.
					if rule.Protocol != "" && rule.Protocol != "tcp" && rule.Protocol != "any" {
						continue
					}

					// A rule without a port range allows all TCP ports.
					if rule.PortRangeMin == 0 && rule.PortRangeMax == 0 {
						for _, port := range portsToCheck {
							allowedPorts[port] = true
						}

						continue
					}

					// Mark every checked port as allowed when it matches the rule range.
					for _, port := range portsToCheck {
						if port >= rule.PortRangeMin && port <= rule.PortRangeMax {
							allowedPorts[port] = true
						}
					}
				}
			}

			return true, nil
		})
		if err != nil {
			return nil, fmt.Errorf("could not list security group %q: %w", sgID, err)
		}
	}

	// A port is restricted when no matching ingress rule allows it.
	for _, port := range portsToCheck {
		if !allowedPorts[port] {
			restrictedPortsList = append(restrictedPortsList, fmt.Sprint(port))
		}
	}

	return restrictedPortsList, nil
}
