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
	"context"
	"log/slog"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleServer creates a virtual machine resource based on the CSC Hub Ontology
func (d *ionosCollector) handleServer(server ionoscloud.Server, dc ionoscloud.Datacenter) (ontology.IsResource, error) {
	l, _, err := d.client.LabelsApi.
		DatacentersServersLabelsGet(context.Background(), pointer.Deref(dc.GetId()), pointer.Deref(server.GetId())).
		Execute()
	if err != nil {
		log.Error("error getting labels for server", slog.String("id", pointer.Deref(server.Id)), tint.Err(err))
	}

	r := &ontology.VirtualMachine{
		Id:                  server.Id,
		Name:                server.Properties.Name,
		CreationTime:        timestamppb.New(pointer.Deref(server.Metadata.GetCreatedDate())),
		GeoLocation:         &ontology.GeoLocation{Region: dc.Properties.GetLocation()},
		Labels:              labels(l),
		ParentId:            dc.GetId(),
		Raw:                 new(collector.Raw(server, dc)),
		BlockStorageIds:     getBlockStorageIds(server),
		NetworkInterfaceIds: getNetworkInterfaceIds(server),
		ActivityLogging: &ontology.ActivityLogging{
			Enabled: new(true), // activity logging is always enabled for IONOS Cloud servers
		},
	}

	return r, nil
}

// handleBlockStorage creates a block storage resource based on the CSC Hub Ontology
func (d *ionosCollector) handleBlockStorage(blockStorage ionoscloud.Volume, dc ionoscloud.Datacenter) (ontology.IsResource, error) {
	l, _, err := d.client.LabelsApi.
		DatacentersServersLabelsGet(context.Background(), pointer.Deref(dc.GetId()), pointer.Deref(blockStorage.GetId())).
		Execute()
	if err != nil {
		log.Error("error getting labels for block storage", slog.String("id", pointer.Deref(blockStorage.Id)), tint.Err(err))
	}

	r := &ontology.BlockStorage{
		Id:           blockStorage.Id,
		Name:         blockStorage.Properties.Name,
		CreationTime: timestamppb.New(pointer.Deref(blockStorage.Metadata.GetCreatedDate())),
		GeoLocation:  &ontology.GeoLocation{Region: dc.Properties.GetLocation()},
		Labels:       labels(l),
		ParentId:     dc.GetId(),
		Raw:          new(collector.Raw(blockStorage, dc)),
	}

	return r, nil
}

// handleLoadBalancer creates a load balancer resource based on the CSC Hub Ontology
func (d *ionosCollector) handleLoadBalancer(loadBalancer ionoscloud.Loadbalancer, dc ionoscloud.Datacenter) (ontology.IsResource, error) {
	l, _, err := d.client.LabelsApi.
		DatacentersServersLabelsGet(context.Background(), pointer.Deref(dc.GetId()), pointer.Deref(loadBalancer.GetId())).
		Execute()
	if err != nil {
		log.Error("error getting labels for load balancer", slog.String("id", pointer.Deref(loadBalancer.Id)), tint.Err(err))
	}

	r := &ontology.LoadBalancer{
		Id:           loadBalancer.Id,
		Name:         loadBalancer.Properties.Name,
		CreationTime: timestamppb.New(pointer.Deref(loadBalancer.Metadata.GetCreatedDate())),
		GeoLocation:  &ontology.GeoLocation{Region: dc.Properties.Location},
		Labels:       labels(l),
		ParentId:     dc.GetId(),
		Raw:          new(collector.Raw(loadBalancer, dc)),
	}

	return r, nil
}
