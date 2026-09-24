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

func Test_ionosCollector_handleDatacenter(t *testing.T) {
	d := newMockIonosCollector(newMockSender())

	testTime, err := time.Parse(time.RFC3339, testdata.MockIonosCreationTime)
	assert.NoError(t, err)

	dc := ionoscloud.Datacenter{
		Id: new(testdata.MockIonosDatacenterID1),
		Properties: &ionoscloud.DatacenterProperties{
			Name:        new(testdata.MockIonosDatacenterName1),
			Description: new(testdata.MockIonosDatacenterDescription1),
			Location:    new(testdata.MockIonosDatacenterLocation1),
		},
		Metadata: &ionoscloud.DatacenterElementMetadata{
			CreatedDate: &ionoscloud.IonosTime{Time: testTime},
		},
	}

	got, err := d.handleDatacenter(dc)

	assert.NoError(t, err)

	gotNew := got.(*ontology.Account)
	assert.NotEmpty(t, gotNew.GetRaw())
	gotNew.Raw = new("")

	want := &ontology.Account{
		Id:           new(testdata.MockIonosDatacenterID1),
		Name:         new(testdata.MockIonosDatacenterName1),
		Description:  new(testdata.MockIonosDatacenterDescription1),
		CreationTime: timestamppb.New(testTime),
		GeoLocation: &ontology.GeoLocation{
			Region: new(testdata.MockIonosDatacenterLocation1),
		},
		Labels: map[string]string{"label1": "value1"},
		Raw:    new(""),
	}
	assert.Equal(t, want, gotNew)
}
