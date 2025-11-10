package openstack

import (
	"confirmate.io/core/api/ontology"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

// discoverServer discovers server
func (d *openstackDiscovery) discoverServer() (list []ontology.IsResource, err error) {
	var opts servers.ListOptsBuilder = &servers.ListOpts{}
	list, err = genericList(d, d.computeClient, servers.List, d.handleServer, servers.ExtractServers, opts)

	return
}
