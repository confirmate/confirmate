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

package testdata

const (
	// location
	MockIonosDatacenterLocation1 = "us/las"
	MockIonosDatacenterLocation2 = "us/dfw"

	// volume
	MockIonosVolumeID1   = "39d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosVolumeID2   = "49d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosVolumeName1 = "Mock IONOS Block Storage 1"
	MockIonosVolumeName2 = "Mock IONOS Block Storage 2"

	// datacenter
	MockIonosDatacenterID1          = "99d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosDatacenterID2          = "a9d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosDatacenterName1        = "Mock IONOS Datacenter 1"
	MockIonosDatacenterName2        = "Mock IONOS Datacenter 2"
	MockIonosDatacenterDescription1 = "Mock IONOS Datacenter 1 Description"
	MockIonosDatacenterDescription2 = "Mock IONOS Datacenter 2 Description"

	// server
	MockIonosVMID1   = "79d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosVMID2   = "89d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosVMName1 = "Mock IONOS VM1"
	MockIonosVMName2 = "Mock IONOS VM2"

	// load balancer
	MockIonosLoadBalancerID1   = "b9d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosLoadBalancerName1 = "Mock IONOS Load Balancer 1"
	MockIonosLoadBalancerID2   = "c9d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosLoadBalancerName2 = "Mock IONOS Load Balancer 2"

	// network
	MockIonosNicID1   = "59d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosNicID2   = "69d85e98-c3da-11ed-afa1-0242ac120002"
	MockIonosNicName1 = "Mock IONOS NIC 1"
	MockIonosNicName2 = "Mock IONOS NIC 2"

	// creation time (RFC3339, used in mocked IONOS API responses)
	MockIonosCreationTime = "2023-02-27T10:00:00Z"
)
