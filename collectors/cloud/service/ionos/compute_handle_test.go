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

func mockDatacenterForHandle() ionoscloud.Datacenter {
	return ionoscloud.Datacenter{
		Id: new(testdata.MockIonosDatacenterID1),
		Properties: &ionoscloud.DatacenterProperties{
			Location: new(testdata.MockIonosDatacenterLocation1),
		},
	}
}

func Test_ionosCollector_handleServer(t *testing.T) {
	d := newMockIonosCollector(newMockSender())

	testTime, err := time.Parse(time.RFC3339, testdata.MockIonosCreationTime)
	assert.NoError(t, err)

	dc := mockDatacenterForHandle()

	server := ionoscloud.Server{
		Id: new(testdata.MockIonosVMID1),
		Properties: &ionoscloud.ServerProperties{
			Name: new(testdata.MockIonosVMName1),
		},
		Metadata: &ionoscloud.DatacenterElementMetadata{
			CreatedDate: &ionoscloud.IonosTime{Time: testTime},
		},
		Entities: &ionoscloud.ServerEntities{
			Volumes: &ionoscloud.AttachedVolumes{Items: &[]ionoscloud.Volume{}},
			Nics:    &ionoscloud.Nics{Items: &[]ionoscloud.Nic{}},
		},
	}

	got, err := d.handleServer(server, dc)

	assert.NoError(t, err)

	gotNew := got.(*ontology.VirtualMachine)
	assert.NotEmpty(t, gotNew.GetRaw())
	gotNew.Raw = new("")

	want := &ontology.VirtualMachine{
		Id:           new(testdata.MockIonosVMID1),
		Name:         new(testdata.MockIonosVMName1),
		CreationTime: timestamppb.New(testTime),
		GeoLocation: &ontology.GeoLocation{
			Region: new(testdata.MockIonosDatacenterLocation1),
		},
		Labels:   map[string]string{"label1": "value1"},
		ParentId: new(testdata.MockIonosDatacenterID1),
		Raw:      new(""),
		ActivityLogging: &ontology.ActivityLogging{
			Enabled: new(true),
		},
	}
	assert.Equal(t, want, gotNew)
}

func Test_ionosCollector_handleBlockStorage(t *testing.T) {
	d := newMockIonosCollector(newMockSender())

	testTime, err := time.Parse(time.RFC3339, testdata.MockIonosCreationTime)
	assert.NoError(t, err)

	dc := mockDatacenterForHandle()

	volume := ionoscloud.Volume{
		Id: new(testdata.MockIonosVolumeID1),
		Properties: &ionoscloud.VolumeProperties{
			Name: new(testdata.MockIonosVolumeName1),
		},
		Metadata: &ionoscloud.DatacenterElementMetadata{
			CreatedDate: &ionoscloud.IonosTime{Time: testTime},
		},
	}

	got, err := d.handleBlockStorage(volume, dc)

	assert.NoError(t, err)

	gotNew := got.(*ontology.BlockStorage)
	assert.NotEmpty(t, gotNew.GetRaw())
	gotNew.Raw = new("")

	want := &ontology.BlockStorage{
		Id:           new(testdata.MockIonosVolumeID1),
		Name:         new(testdata.MockIonosVolumeName1),
		CreationTime: timestamppb.New(testTime),
		GeoLocation: &ontology.GeoLocation{
			Region: new(testdata.MockIonosDatacenterLocation1),
		},
		Labels:   map[string]string{"label1": "value1"},
		ParentId: new(testdata.MockIonosDatacenterID1),
		Raw:      new(""),
	}
	assert.Equal(t, want, gotNew)
}

func Test_ionosCollector_handleLoadBalancer(t *testing.T) {
	d := newMockIonosCollector(newMockSender())

	testTime, err := time.Parse(time.RFC3339, testdata.MockIonosCreationTime)
	assert.NoError(t, err)

	dc := mockDatacenterForHandle()

	lb := ionoscloud.Loadbalancer{
		Id: new(testdata.MockIonosLoadBalancerID1),
		Properties: &ionoscloud.LoadbalancerProperties{
			Name: new(testdata.MockIonosLoadBalancerName1),
		},
		Metadata: &ionoscloud.DatacenterElementMetadata{
			CreatedDate: &ionoscloud.IonosTime{Time: testTime},
		},
	}

	got, err := d.handleLoadBalancer(lb, dc)

	assert.NoError(t, err)

	gotNew := got.(*ontology.LoadBalancer)
	assert.NotEmpty(t, gotNew.GetRaw())
	gotNew.Raw = new("")

	want := &ontology.LoadBalancer{
		Id:           new(testdata.MockIonosLoadBalancerID1),
		Name:         new(testdata.MockIonosLoadBalancerName1),
		CreationTime: timestamppb.New(testTime),
		GeoLocation: &ontology.GeoLocation{
			Region: new(testdata.MockIonosDatacenterLocation1),
		},
		ParentId: new(testdata.MockIonosDatacenterID1),
		Raw:      new(""),
	}
	assert.Equal(t, want, gotNew)
}
