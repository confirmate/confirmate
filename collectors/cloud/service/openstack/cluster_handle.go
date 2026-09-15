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

	"github.com/gophercloud/gophercloud/v2/openstack/containerinfra/v1/clusters"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleCluster creates a container resource based on the CSC Hub Ontology
func (d *openstackCollector) handleCluster(cluster *clusters.Cluster) (ontology.IsResource, error) {
	r := &ontology.ContainerOrchestration{
		Id:           cluster.UUID,
		Name:         cluster.Name,
		CreationTime: timestamppb.New(cluster.CreatedAt),
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		Labels:   cluster.Labels,
		ParentId: new(cluster.ProjectID),
		Raw:      collector.Raw(cluster),
	}

	log.Info("Adding cluster", slog.String("name", cluster.Name))

	return r, nil
}
