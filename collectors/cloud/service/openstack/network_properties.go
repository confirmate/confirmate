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

// getRestrictedPorts returns the list of ports restricted by the security groups' ingress rules.
func (d *openstackCollector) getRestrictedPorts(portSecurityGroup []string) []string {
	var restrictedPortsList []string

	for _, sgID := range portSecurityGroup {
		pager := groups.List(d.clients.networkClient, groups.ListOpts{
			ID: sgID,
		})

		err := pager.EachPage(context.Background(), func(_ context.Context, page pagination.Page) (bool, error) {
			sgList, err := groups.ExtractGroups(page)
			if err != nil {
				return false, err
			}

			for _, sg := range sgList {
				for _, rule := range sg.Rules {
					if rule.Direction != "ingress" {
						continue
					}
					if rule.Protocol != "tcp" && rule.Protocol != "udp" {
						continue
					}

					switch {
					case rule.PortRangeMin == 0 && rule.PortRangeMax == 0:
						// If no port range is specified, it means all ports are allowed
						restrictedPortsList = append(restrictedPortsList, "all")
					case rule.PortRangeMin == rule.PortRangeMax:
						restrictedPortsList = append(restrictedPortsList, fmt.Sprint(rule.PortRangeMin))
					default:
						for port := rule.PortRangeMin; port <= rule.PortRangeMax; port++ {
							restrictedPortsList = append(restrictedPortsList, fmt.Sprint(port))
						}
					}
				}
			}

			return true, nil
		})
		if err != nil {
			continue
		}
	}

	return restrictedPortsList
}
