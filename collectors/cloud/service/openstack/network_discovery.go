package openstack

import (
	"confirmate.io/collectors/cloud/api/ontology"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
)

// discoverNetworkInterfaces discovers network interfaces
func (d *openstackDiscovery) discoverNetworkInterfaces() (list []ontology.IsResource, err error) {
	var opts networks.ListOptsBuilder = &networks.ListOpts{}
	list, err = genericList(d, d.networkClient, networks.List, d.handleNetworkInterfaces, networks.ExtractNetworks, opts)

	return
}
