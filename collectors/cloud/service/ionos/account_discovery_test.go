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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_ionosCollector_collectDatacenters(t *testing.T) {
	d := newMockIonosCollector(newMockSender())

	dc, list, err := d.collectDatacenters()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, 2, len(*dc.Items))

	got0 := list[0].(*ontology.Account)
	assert.NotEmpty(t, got0.GetRaw())
	got0.Raw = new("")

	want := &ontology.Account{
		Id:           new(testdata.MockIonosDatacenterID1),
		Name:         new(testdata.MockIonosDatacenterName1),
		Description:  new(testdata.MockIonosDatacenterDescription1),
		CreationTime: timestamppb.New(time.Time{}),
		GeoLocation: &ontology.GeoLocation{
			Region: new(testdata.MockIonosDatacenterLocation1),
		},
		Labels: map[string]string{"label1": "value1"},
		Raw:    new(""),
	}
	assert.Equal(t, want, got0)
}

func Test_ionosCollector_collectDatacenters_error(t *testing.T) {
	d := newMockIonosCollector(newMockErrorSender())

	_, _, err := d.collectDatacenters()

	assert.Error(t, err)
}
