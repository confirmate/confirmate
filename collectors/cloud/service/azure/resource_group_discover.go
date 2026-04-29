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
	"context"
	"fmt"
	"log/slog"
	"strings"

	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// collectResourceGroups collects resource groups and cloud account
func (d *azureCollector) collectResourceGroups() (list []ontology.IsResource, err error) {
	// initialize client
	if err := d.initResourceGroupsClient(); err != nil {
		return nil, err
	}

	// Build an account as the most top-level item. Our subscription will serve as the account
	acc := d.handleSubscription(d.sub)

	list = append(list, acc)

	listPager := d.clients.rgClient.NewListPager(&armresources.ResourceGroupsClientListOptions{})
	for listPager.More() {
		page, err := listPager.NextPage(context.TODO())
		if err != nil {
			err = fmt.Errorf("%s: %v", ErrGettingNextPage, err)
			return nil, err
		}

		for _, rg := range page.Value {
			// If we are scoped to one resource group, we can skip the rest of the groups. Resource group names are
			// case-insensitive
			if d.rg != nil && !strings.EqualFold(pointer.Deref(rg.Name), pointer.Deref(d.rg)) {
				continue
			}

			r := d.handleResourceGroup(rg)

			log.Info("Adding resource group", slog.String("resource group", r.GetName()))

			list = append(list, r)
		}
	}

	return
}
