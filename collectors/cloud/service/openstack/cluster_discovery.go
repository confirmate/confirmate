package openstack

import (
	"confirmate.io/core/api/ontology"

	"github.com/gophercloud/gophercloud/v2/openstack/containerinfra/v1/clusters"
)

// collectCluster collects OpenStack clusters and returns a list of resources
func (d *openstackCollector) collectCluster() (list []ontology.IsResource, err error) {
	var opts clusters.ListOptsBuilder = &clusters.ListOpts{}
	list, err = genericList(d, d.clusterClient, clusters.List, d.handleCluster, clusters.ExtractClusters, opts)

	return
}
