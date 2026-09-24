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
	"confirmate.io/core/api/ontology"

	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/containers"
)

// collectBlockStorage collects block storages
func (d *openstackCollector) collectBlockStorage() (list []ontology.IsResource, err error) {
	var opts volumes.ListOptsBuilder = &volumes.ListOpts{}
	list, err = genericList(d, d.blockStorageClient, volumes.List, d.handleBlockStorage, volumes.ExtractVolumes, opts)

	return
}

// collectObjectStorage collects object storages
func (d *openstackCollector) collectObjectStorage() (list []ontology.IsResource, err error) {
	var opts containers.ListOptsBuilder = &containers.ListOpts{}
	list, err = genericList(d, d.storageClient, containers.List, d.handleObjectStorage, containers.ExtractInfo, opts)

	return
}

// collectObjectStorageService collects the object storage service resource. It is a no-op if the object storage
// client could not be initialized, e.g. because the deployment does not offer an object storage service.
func (d *openstackCollector) collectObjectStorageService() (list []ontology.IsResource, err error) {
	if d.clients.storageClient == nil {
		return nil, nil
	}

	resource, err := d.handleObjectStorageService()
	if err != nil {
		return nil, err
	}

	list = append(list, resource)

	return
}
