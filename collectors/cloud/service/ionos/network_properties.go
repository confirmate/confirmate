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
	"strconv"

	"confirmate.io/collectors/cloud/internal/pointer"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// getRestrictedPorts returns the list of ports restricted by firewall rules on the given NIC.
func (d *ionosCollector) getRestrictedPorts(nic ionoscloud.Nic) []string {
	var restrictedPortsList []string

	if nic.Entities == nil || nic.Entities.Firewallrules == nil || nic.Entities.Firewallrules.Items == nil {
		return restrictedPortsList
	}

	for _, rule := range *nic.Entities.Firewallrules.Items {
		if rule.Properties == nil {
			continue
		}

		start := pointer.Deref(rule.Properties.PortRangeStart)
		end := pointer.Deref(rule.Properties.PortRangeEnd)

		switch {
		case start == 0 && end == 0:
			// If no port range is specified, it means all ports are allowed
			restrictedPortsList = append(restrictedPortsList, "all")
		case start == end:
			// If the port range is a single port, add that port to the list
			restrictedPortsList = append(restrictedPortsList, strconv.Itoa(int(start)))
		case rule.Properties.PortRangeStart != nil && rule.Properties.PortRangeEnd != nil:
			// If the port range is specified, add each port in the range to the list
			for port := start; port <= end; port++ {
				restrictedPortsList = append(restrictedPortsList, strconv.Itoa(int(port)))
			}
		}
	}

	return restrictedPortsList
}
