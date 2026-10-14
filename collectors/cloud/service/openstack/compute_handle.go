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
	"log/slog"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/core/api/ontology"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleServer creates a virtual machine resource based on the CSC Hub Ontology
func (d *openstackCollector) handleServer(server *servers.Server) (ontology.IsResource, error) {
	var (
		err                error
		bootLogging        *ontology.BootLogging
		activityLogEnabled bool
	)

	// we cannot directly retrieve OS logging information
	// boot logging is logged in the console log
	consoleOutput := servers.ShowConsoleOutput(context.Background(), d.clients.computeClient, server.ID, servers.ShowConsoleOutputOpts{})
	if consoleOutput.Result.Err == nil {
		bootLogging = &ontology.BootLogging{
			Enabled: new(true),
		}
	} else {
		log.Error("Error getting boot logging", tint.Err(consoleOutput.Err))
		// When an error occurs, we assume that boot logging is disabled.
		bootLogging = &ontology.BootLogging{
			Enabled: new(false),
		}
	}

	// OpenStack does not provide a direct way to check if activity logging is enabled for a server. As a
	// heuristic, we treat activity logging as enabled if the Telemetry service (event or metering endpoints)
	// is available.
	eventEndpoint, errEvent := d.clients.provider.EndpointLocator(gophercloud.EndpointOpts{
		Type:   "event",
		Region: d.region,
	})
	meteringEndpoint, errMetering := d.clients.provider.EndpointLocator(gophercloud.EndpointOpts{
		Type:   "metering",
		Region: d.region,
	})
	activityLogEnabled = (errEvent == nil && eventEndpoint != "") || (errMetering == nil && meteringEndpoint != "")

	r := &ontology.VirtualMachine{
		Id:           new(server.ID),
		Name:         new(server.Name),
		CreationTime: timestamppb.New(server.Created),
		GeoLocation: &ontology.GeoLocation{
			Region: new(d.region),
		},
		Labels:   labels(server.Tags),
		ParentId: new(server.TenantID),
		Raw:      new(collector.Raw(server)),
		// MalwareProtection: &ontology.MalwareProtection{},
		BootLogging: bootLogging,
		// AutomaticUpdates:  &ontology.AutomaticUpdates{}, // Information not available in OpenStack, so we leave it empty
		ActivityLogging: &ontology.ActivityLogging{
			Enabled: new(activityLogEnabled),
		},
		InternetAccessibleEndpoint: new(hasFloatingIP(server.Addresses)),
	}

	// Get attached block storage IDs
	for _, v := range server.AttachedVolumes {
		r.BlockStorageIds = append(r.BlockStorageIds, v.ID)
	}

	// Get attached network interface IDs
	r.NetworkInterfaceIds, err = d.getAttachedNetworkInterfaces(server.ID)
	if err != nil {
		return nil, fmt.Errorf("could not collect attached network interfaces: %w", err)
	}

	log.Info("Adding server", slog.String("name", server.Name))

	return r, nil
}

// hasFloatingIP reports whether any address of a server is a floating IP, as indicated by the
// "OS-EXT-IPS:type" attribute of the Nova addresses. Floating IPs are the usual way to expose a server to
// the internet. Fixed addresses on provider networks that are publicly routable are not detected, and
// servers without any address (e.g., shelved ones) are reported as not accessible.
func hasFloatingIP(addresses map[string]any) bool {
	for _, network := range addresses {
		entries, ok := network.([]any)
		if !ok {
			continue
		}

		for _, entry := range entries {
			addr, ok := entry.(map[string]any)
			if ok && addr["OS-EXT-IPS:type"] == "floating" {
				return true
			}
		}
	}

	return false
}
