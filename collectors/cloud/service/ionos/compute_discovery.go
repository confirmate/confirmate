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
	"fmt"
	"log/slog"

	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// collectServers lists all virtual machines and corresponding network interfaces in the given datacenter and
// returns them as a list of ontology resources.
func (d *ionosCollector) collectServers(dc ionoscloud.Datacenter) (list []ontology.IsResource, err error) {
	// Depth(5) returns all available properties, including nested entities such as NICs and volumes.
	servers, _, err := d.client.ServersApi.DatacentersServersGet(context.Background(), pointer.Deref(dc.Id)).Depth(5).Execute()
	if err != nil {
		return nil, fmt.Errorf("could not list servers for datacenter %s: %w", pointer.Deref(dc.Id), err)
	}

	for _, server := range pointer.Deref(servers.Items) {
		r, err := d.handleServer(server, dc)
		if err != nil {
			return nil, fmt.Errorf("could not handle server %s: %w", pointer.Deref(server.Id), err)
		}

		log.Info("Adding server", slog.String("id", r.GetId()))

		list = append(list, r)

		for _, nic := range pointer.Deref(server.Entities.GetNics().GetItems()) {
			networkInterface, err := d.handleNetworkInterface(nic, dc)
			if err != nil {
				return nil, fmt.Errorf("could not handle network interfaces: %w", err)
			}

			log.Info("Adding network interface", slog.String("id", pointer.Deref(nic.GetId())))

			list = append(list, networkInterface)
		}
	}

	return list, nil
}

// collectBlockStorages lists all block storages in the given datacenter and returns them as a list of ontology
// resources.
func (d *ionosCollector) collectBlockStorages(dc ionoscloud.Datacenter) (list []ontology.IsResource, err error) {
	blockStorages, _, err := d.client.VolumesApi.DatacentersVolumesGet(context.Background(), pointer.Deref(dc.Id)).Depth(1).Execute()
	if err != nil {
		return nil, fmt.Errorf("could not list block storages for datacenter %s: %w", pointer.Deref(dc.Id), err)
	}

	for _, blockStorage := range pointer.Deref(blockStorages.Items) {
		r, err := d.handleBlockStorage(blockStorage, dc)
		if err != nil {
			return nil, fmt.Errorf("could not handle block storage %s: %w", pointer.Deref(blockStorage.Id), err)
		}

		log.Info("Adding block storage", slog.String("id", r.GetId()))

		list = append(list, r)
	}

	return list, nil
}

// collectLoadBalancers lists all load balancers and corresponding network interfaces in the given datacenter and
// returns them as a list of ontology resources.
func (d *ionosCollector) collectLoadBalancers(dc ionoscloud.Datacenter) (list []ontology.IsResource, err error) {
	// Depth(5) returns all available properties, including nested entities such as balanced NICs.
	loadBalancers, _, err := d.client.LoadBalancersApi.DatacentersLoadbalancersGet(context.Background(), pointer.Deref(dc.Id)).Depth(5).Execute()
	if err != nil {
		return nil, fmt.Errorf("could not list load balancers for datacenter %s: %w", pointer.Deref(dc.Id), err)
	}

	for _, loadBalancer := range pointer.Deref(loadBalancers.Items) {
		r, err := d.handleLoadBalancer(loadBalancer, dc)
		if err != nil {
			return nil, fmt.Errorf("could not handle load balancer %s: %w", pointer.Deref(loadBalancer.Id), err)
		}

		log.Info("Adding load balancer", slog.String("id", r.GetId()))

		list = append(list, r)

		for _, nic := range pointer.Deref(loadBalancer.Entities.GetBalancednics().GetItems()) {
			networkInterface, err := d.handleNetworkInterface(nic, dc)
			if err != nil {
				return nil, fmt.Errorf("could not handle network interfaces: %w", err)
			}

			log.Info("Adding network interface", slog.String("id", pointer.Deref(nic.GetId())))

			list = append(list, networkInterface)
		}
	}

	return list, nil
}
