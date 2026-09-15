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

	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleServer creates a virtual machine resource based on the CSC Hub Ontology
func (d *openstackCollector) handleServer(server *servers.Server) (ontology.IsResource, error) {
	var (
		err         error
		bootLogging *ontology.BootLogging
	)

	// we cannot directly retrieve OS logging information
	// boot logging is logged in the console log
	consoleOutput := servers.ShowConsoleOutput(context.Background(), d.clients.computeClient, server.ID, servers.ShowConsoleOutputOpts{})
	if consoleOutput.Result.Err == nil {
		bootLogging = &ontology.BootLogging{
			Enabled: true,
		}
	} else {
		log.Error("Error getting boot logging", tint.Err(consoleOutput.Err))
		// When an error occurs, we assume that boot logging is disabled.
		bootLogging = &ontology.BootLogging{
			Enabled: false,
		}
	}

	r := &ontology.VirtualMachine{
		Id:           server.ID,
		Name:         server.Name,
		CreationTime: timestamppb.New(server.Created),
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		Labels:            labels(server.Tags),
		ParentId:          new(server.TenantID),
		Raw:               collector.Raw(server),
		MalwareProtection: &ontology.MalwareProtection{},
		BootLogging:       bootLogging,
		AutomaticUpdates:  &ontology.AutomaticUpdates{},
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
