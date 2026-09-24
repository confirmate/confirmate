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
	"context"
	"fmt"
	"log/slog"

	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// collectDatacenters lists all datacenters in the IONOS Cloud and returns them as a list of ontology resources.
func (d *ionosCollector) collectDatacenters() (*ionoscloud.Datacenters, []ontology.IsResource, error) {
	var list []ontology.IsResource

	items, err := paginate(func(offset int32) ([]ionoscloud.Datacenter, error) {
		dc, _, err := d.client.DataCentersApi.DatacentersGet(context.Background()).Depth(1).Offset(offset).Limit(pageLimit).Execute()
		if err != nil {
			return nil, err
		}
		return pointer.Deref(dc.Items), nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not list datacenters: %w", err)
	}

	for _, datacenter := range items {
		r, err := d.handleDatacenter(datacenter)
		if err != nil {
			return nil, nil, fmt.Errorf("could not handle datacenter %s: %w", pointer.Deref(datacenter.Id), err)
		}

		log.Info("Adding datacenter", slog.String("id", r.GetId()))

		list = append(list, r)
	}

	return &ionoscloud.Datacenters{Items: &items}, list, nil
}
