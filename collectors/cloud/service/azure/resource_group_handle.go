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

package azure

import (
	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// handleResourceGroup returns a [ontology.ResourceGroup] out of an existing [armresources.ResourceGroup].
func (d *azureCollector) handleResourceGroup(rg *armresources.ResourceGroup) ontology.IsResource {
	return &ontology.ResourceGroup{
		Id:          resourceID(rg.ID),
		Name:        pointer.Deref(rg.Name),
		GeoLocation: location(rg.Location),
		Labels:      labels(rg.Tags),
		ParentId:    d.sub.ID,
		Raw:         collector.Raw(rg),
	}
}

// handleSubscription returns a [ontology.Account] out of an existing [armsubscription.Subscription].
func (d *azureCollector) handleSubscription(s *armsubscription.Subscription) *ontology.Account {
	return &ontology.Account{
		Id:           resourceID(s.ID),
		Name:         pointer.Deref(s.DisplayName),
		CreationTime: nil, // subscriptions do not have a creation date
		GeoLocation:  nil, // subscriptions are global
		Labels:       nil, // subscriptions do not have labels,
		ParentId:     nil, // subscriptions are the top-most item and have no parent,
		Raw:          collector.Raw(s),
	}
}
