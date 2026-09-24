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
	"fmt"
	"net/http"
	"testing"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
)

func Test_openstackCollector_collectIdentity(t *testing.T) {
	fakeServer := testhelper.SetupHTTP()
	defer fakeServer.Teardown()

	fakeServer.Mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "GET")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
{
  "users": [
    {
      "id": "test-user-id",
      "name": "test-user",
      "default_project_id": "test-project-id",
      "enabled": true
    }
  ],
  "links": {"next": null, "previous": null}
}
`)
	})

	d := &openstackCollector{
		clients: clients{
			identityClient: client.ServiceClient(fakeServer),
		},
		region:  "test region",
		project: &project{},
	}

	gotList, err := d.collectIdentity()

	assert.NoError(t, err)
	assert.Equal(t, 1, len(gotList))

	got0 := gotList[0].(*ontology.Identity)
	assert.NotEmpty(t, got0.GetRaw())
	got0.Raw = new("")

	want := &ontology.Identity{
		Id:   new("test-user-id"),
		Name: new("test-user"),
		GeoLocation: &ontology.GeoLocation{
			Region: new("test region"),
		},
		ParentId:              new("test-project-id"),
		Activated:             new(true),
		Raw:                   new(""),
		DisablePasswordPolicy: new(true),
	}
	assert.Equal(t, want, got0)
}
