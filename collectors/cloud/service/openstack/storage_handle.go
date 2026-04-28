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
	"log/slog"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleBlockStorage creates a block storage resource based on the CSC Hub Ontology
func (d *openstackCollector) handleBlockStorage(volume *volumes.Volume) (ontology.IsResource, error) {
	// Get Name, if exits, otherwise take the ID
	name := volume.Name
	if volume.Name == "" {
		name = volume.ID
	}

	r := &ontology.BlockStorage{
		Id:           volume.ID,
		Name:         name,
		Description:  volume.Description,
		CreationTime: timestamppb.New(volume.CreatedAt),
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		ParentId: util.Ref(getParentID(volume)),
		Labels:   map[string]string{}, // Not available
		Raw:      collector.Raw(volume),
	}

	log.Info("Adding block storage", slog.String("name", volume.Name))

	return r, nil
}
