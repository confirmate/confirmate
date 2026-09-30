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
	"confirmate.io/collectors/cloud/internal/pointer"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// getBlockStorageIds lists the block storage IDs attached to the given server
func getBlockStorageIds(server ionoscloud.Server) (blockStorageIds []string) {
	for _, volume := range pointer.Deref(server.Entities.GetVolumes().GetItems()) {
		blockStorageIds = append(blockStorageIds, pointer.Deref(volume.GetId()))
	}

	return blockStorageIds
}

// getNetworkInterfaceIds lists the network interface IDs attached to the given server
func getNetworkInterfaceIds(server ionoscloud.Server) (networkInterfaceIds []string) {
	for _, nic := range pointer.Deref(server.Entities.GetNics().GetItems()) {
		networkInterfaceIds = append(networkInterfaceIds, pointer.Deref(nic.GetId()))
	}

	return networkInterfaceIds
}
