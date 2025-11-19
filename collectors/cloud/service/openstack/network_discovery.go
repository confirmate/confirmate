package openstack

import (
	"confirmate.io/core/api/ontology"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
)

// collectNetworkInterfaces collects network interfaces
func (d *openstackCollector) collectNetworkInterfaces() (list []ontology.IsResource, err error) {
	var opts networks.ListOptsBuilder = &networks.ListOpts{}
	list, err = genericList(d, d.networkClient, networks.List, d.handleNetworkInterfaces, networks.ExtractNetworks, opts)

	return
}
