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
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func mockDatacenter1(t *testing.T) ionoscloud.Datacenter {
	d := newMockIonosCollector(newMockSender())

	dc, _, err := d.collectDatacenters()
	assert.NoError(t, err)

	for _, item := range *dc.Items {
		if item.GetId() != nil && *item.GetId() == testdata.MockIonosDatacenterID1 {
			return item
		}
	}

	t.Fatalf("could not find mock datacenter %s", testdata.MockIonosDatacenterID1)
	return ionoscloud.Datacenter{}
}

func Test_ionosCollector_collectServers(t *testing.T) {
	d := newMockIonosCollector(newMockSender())
	dc := mockDatacenter1(t)

	list, err := d.collectServers(dc)

	assert.NoError(t, err)
	// 1 server + 1 network interface
	assert.Equal(t, 2, len(list))

	got0 := list[0].(*ontology.VirtualMachine)
	assert.Equal(t, testdata.MockIonosVMID1, *got0.Id)
	assert.Equal(t, testdata.MockIonosVMName1, *got0.Name)
	assert.Equal(t, []string{testdata.MockIonosVolumeID1}, got0.BlockStorageIds)
	assert.Equal(t, []string{testdata.MockIonosNicID1}, got0.NetworkInterfaceIds)

	got1 := list[1].(*ontology.NetworkInterface)
	assert.Equal(t, testdata.MockIonosNicID1, *got1.Id)
}

func Test_ionosCollector_collectBlockStorages(t *testing.T) {
	d := newMockIonosCollector(newMockSender())
	dc := mockDatacenter1(t)

	list, err := d.collectBlockStorages(dc)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))

	got0 := list[0].(*ontology.BlockStorage)
	assert.Equal(t, testdata.MockIonosVolumeID1, *got0.Id)
	assert.Equal(t, testdata.MockIonosVolumeName1, *got0.Name)
}

func Test_ionosCollector_collectLoadBalancers(t *testing.T) {
	d := newMockIonosCollector(newMockSender())
	dc := mockDatacenter1(t)

	list, err := d.collectLoadBalancers(dc)

	assert.NoError(t, err)
	// 1 load balancer + 1 network interface
	assert.Equal(t, 2, len(list))

	got0 := list[0].(*ontology.LoadBalancer)
	assert.Equal(t, testdata.MockIonosLoadBalancerID1, *got0.Id)
	assert.Equal(t, testdata.MockIonosLoadBalancerName1, *got0.Name)
}
