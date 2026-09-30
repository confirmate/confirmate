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
	"time"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_ionosCollector_handleNetworkInterface(t *testing.T) {
	d := &ionosCollector{}

	testTime, err := time.Parse(time.RFC3339, testdata.MockIonosCreationTime)
	assert.NoError(t, err)

	dc := ionoscloud.Datacenter{
		Id: new(testdata.MockIonosDatacenterID1),
		Properties: &ionoscloud.DatacenterProperties{
			Location: new(testdata.MockIonosDatacenterLocation1),
		},
	}

	nic := ionoscloud.Nic{
		Id: new(testdata.MockIonosNicID1),
		Properties: &ionoscloud.NicProperties{
			Name:           new(testdata.MockIonosNicName1),
			FirewallActive: new(true),
		},
		Metadata: &ionoscloud.DatacenterElementMetadata{
			CreatedDate: &ionoscloud.IonosTime{Time: testTime},
		},
	}

	got, err := d.handleNetworkInterface(nic, dc)

	assert.NoError(t, err)

	gotNew := got.(*ontology.NetworkInterface)
	assert.NotEmpty(t, gotNew.GetRaw())
	gotNew.Raw = new("")

	want := &ontology.NetworkInterface{
		Id:           new(testdata.MockIonosNicID1),
		Name:         new(testdata.MockIonosNicName1),
		CreationTime: timestamppb.New(testTime),
		GeoLocation: &ontology.GeoLocation{
			Region: new(testdata.MockIonosDatacenterLocation1),
		},
		ParentId: new(testdata.MockIonosDatacenterID1),
		Raw:      new(""),
		AccessRestriction: &ontology.AccessRestriction{
			Type: &ontology.AccessRestriction_L3Firewall{
				L3Firewall: &ontology.L3Firewall{
					Enabled: new(true),
				},
			},
		},
	}
	assert.Equal(t, want, gotNew)
}
