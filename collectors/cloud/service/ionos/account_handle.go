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
	"log/slog"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleDatacenter creates an account resource based on the CSC Hub Ontology
func (d *ionosCollector) handleDatacenter(dc ionoscloud.Datacenter) (ontology.IsResource, error) {
	l, _, err := d.client.LabelsApi.
		DatacentersLabelsGet(context.Background(), pointer.Deref(dc.Id)).
		Execute()
	if err != nil {
		log.Error("error getting labels for datacenter", slog.String("id", pointer.Deref(dc.Id)), tint.Err(err))
	}

	r := &ontology.Account{
		Id:           dc.Id,
		Name:         dc.Properties.Name,
		Description:  dc.Properties.Description,
		CreationTime: timestamppb.New(pointer.Deref(dc.Metadata.GetCreatedDate())),
		GeoLocation: &ontology.GeoLocation{
			Region: dc.Properties.Location,
		},
		Labels: labels(l),
		Raw:    new(collector.Raw(dc)),
	}

	return r, nil
}
