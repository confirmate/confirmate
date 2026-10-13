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

	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumetypes"
)

// getParentID returns the parent ID of a volume.
// The volume can be attached to multiple resources; retrieve the first one that has a serverID assigned.
func getParentID(volume *volumes.Volume) string {
	for _, attach := range volume.Attachments {
		if attach.ServerID != "" {
			return attach.ServerID
		}
	}

	// If no attachment is available, we attach it to the project ID
	return volume.TenantID
}

// getVolumeTypeByName resolves a volume type name to its actual VolumeType object.
// The OpenStack API expects the volume type ID for certain calls (like encryption lookups),
// but volumes only store the type name, so we have to list and match them first.
func (d *openstackCollector) getVolumeTypeByName(typeName string) (*volumetypes.VolumeType, error) {
	// List all available volume types
	allPages, err := volumetypes.List(d.clients.blockStorageClient, volumetypes.ListOpts{}).AllPages(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list volume types: %w", err)
	}

	volumeTypes, err := volumetypes.ExtractVolumeTypes(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract volume types: %w", err)
	}

	// Find the volume type that matches the given name
	for _, vt := range volumeTypes {
		if vt.Name == typeName {
			return &vt, nil
		}
	}

	return nil, fmt.Errorf("volume type '%s' not found", typeName)
}
