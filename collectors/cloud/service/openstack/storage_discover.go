package openstack

import (
	"confirmate.io/core/api/ontology"

	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
)

// discoverBlockStorage discovers block storages
func (d *openstackDiscovery) discoverBlockStorage() (list []ontology.IsResource, err error) {
	var opts volumes.ListOptsBuilder = &volumes.ListOpts{}
	list, err = genericList(d, d.storageClient, volumes.List, d.handleBlockStorage, volumes.ExtractVolumes, opts)

	return
}
