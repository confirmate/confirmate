package azure

import (
	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
)

func (d *azureDiscovery) handleInstances(vault *armdataprotection.BackupVaultResource, instance *armdataprotection.BackupInstanceResource) (resource ontology.IsResource, err error) {
	if vault == nil || instance == nil {
		return nil, ErrVaultInstanceIsEmpty
	}

	if *instance.Properties.DataSourceInfo.DatasourceType == "Microsoft.Storage/storageAccounts/blobServices" {
		resource = &ontology.ObjectStorage{
			Id:          resourceID(instance.ID),
			Name:        util.Deref(instance.Name),
			GeoLocation: location(vault.Location),
			ParentId:    resourceGroupID(instance.ID),
			Raw:         cloud.Raw(instance, vault),
		}
	} else if *instance.Properties.DataSourceInfo.DatasourceType == "Microsoft.Compute/disks" {
		resource = &ontology.BlockStorage{
			Id:          resourceID(instance.ID),
			Name:        util.Deref(instance.Name),
			GeoLocation: location(vault.Location),
			ParentId:    resourceGroupID(instance.ID),
			Raw:         cloud.Raw(instance, vault),
		}
	}

	return
}
