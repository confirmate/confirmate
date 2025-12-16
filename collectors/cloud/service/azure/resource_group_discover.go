package azure

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

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
			if d.rg != nil && !strings.EqualFold(util.Deref(rg.Name), util.Deref(d.rg)) {
				continue
			}

			r := d.handleResourceGroup(rg)

			log.Info("Adding resource group", slog.String("resource group", r.GetName()))

			list = append(list, r)
		}
	}

	return
}
