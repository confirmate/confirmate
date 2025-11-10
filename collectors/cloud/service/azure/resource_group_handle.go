package azure

import (
	"clouditor.io/clouditor/v2/api/ontology"
	"clouditor.io/clouditor/v2/internal/util"
	cloud "confirmate.io/collectors/cloud/api"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// handleResourceGroup returns a [ontology.ResourceGroup] out of an existing [armresources.ResourceGroup].
func (d *azureDiscovery) handleResourceGroup(rg *armresources.ResourceGroup) ontology.IsResource {
	return &ontology.ResourceGroup{
		Id:          resourceID(rg.ID),
		Name:        util.Deref(rg.Name),
		GeoLocation: location(rg.Location),
		Labels:      labels(rg.Tags),
		ParentId:    d.sub.ID,
		Raw:         cloud.Raw(rg),
	}
}

// handleSubscription returns a [ontology.Account] out of an existing [armsubscription.Subscription].
func (d *azureDiscovery) handleSubscription(s *armsubscription.Subscription) *ontology.Account {
	return &ontology.Account{
		Id:           resourceID(s.ID),
		Name:         util.Deref(s.DisplayName),
		CreationTime: nil, // subscriptions do not have a creation date
		GeoLocation:  nil, // subscriptions are global
		Labels:       nil, // subscriptions do not have labels,
		ParentId:     nil, // subscriptions are the top-most item and have no parent,
		Raw:          cloud.Raw(s),
	}
}
