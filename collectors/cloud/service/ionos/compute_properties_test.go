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
	"testing"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/util/assert"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func Test_getBlockStorageIds(t *testing.T) {
	server := ionoscloud.Server{
		Entities: &ionoscloud.ServerEntities{
			Volumes: &ionoscloud.AttachedVolumes{
				Items: &[]ionoscloud.Volume{
					{Id: new(testdata.MockIonosVolumeID1)},
					{Id: new(testdata.MockIonosVolumeID2)},
				},
			},
		},
	}

	got := getBlockStorageIds(server)

	assert.Equal(t, []string{testdata.MockIonosVolumeID1, testdata.MockIonosVolumeID2}, got)
}

func Test_getNetworkInterfaceIds(t *testing.T) {
	server := ionoscloud.Server{
		Entities: &ionoscloud.ServerEntities{
			Nics: &ionoscloud.Nics{
				Items: &[]ionoscloud.Nic{
					{Id: new(testdata.MockIonosNicID1)},
					{Id: new(testdata.MockIonosNicID2)},
				},
			},
		},
	}

	got := getNetworkInterfaceIds(server)

	assert.Equal(t, []string{testdata.MockIonosNicID1, testdata.MockIonosNicID2}, got)
}
