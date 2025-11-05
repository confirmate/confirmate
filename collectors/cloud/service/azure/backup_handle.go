package azure

import (
	"clouditor.io/clouditor/v2/api/discovery"
	"clouditor.io/clouditor/v2/api/ontology"
	"clouditor.io/clouditor/v2/internal/util"
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
			Raw:         discovery.Raw(instance, vault),
		}
	} else if *instance.Properties.DataSourceInfo.DatasourceType == "Microsoft.Compute/disks" {
		resource = &ontology.BlockStorage{
			Id:          resourceID(instance.ID),
			Name:        util.Deref(instance.Name),
			GeoLocation: location(vault.Location),
			ParentId:    resourceGroupID(instance.ID),
			Raw:         discovery.Raw(instance, vault),
		}
	}

	return
}
