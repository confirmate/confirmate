package openstack

import (
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
)

// identityClient returns the identity client if initialized
func (d *openstackDiscovery) identityClient() (client *gophercloud.ServiceClient, err error) {
	if d.clients.identityClient == nil {
		return nil, fmt.Errorf("identity client not initialized")
	}
	return d.clients.identityClient, nil
}

// computeClient returns the compute client if initialized
func (d *openstackDiscovery) computeClient() (client *gophercloud.ServiceClient, err error) {
	if d.clients.computeClient == nil {
		return nil, fmt.Errorf("compute client not initialized")
	}
	return d.clients.computeClient, nil
}

// networkClient returns the network client if initialized
func (d *openstackDiscovery) networkClient() (client *gophercloud.ServiceClient, err error) {
	if d.clients.networkClient == nil {
		return nil, fmt.Errorf("network client not initialized")
	}
	return d.clients.networkClient, nil
}

// storageClient returns the compute client if initialized
func (d *openstackDiscovery) storageClient() (client *gophercloud.ServiceClient, err error) {
	if d.clients.storageClient == nil {
		return nil, fmt.Errorf("storage client not initialized")
	}
	return d.clients.storageClient, nil
}

// clusterClient returns the cluster client if initialized
func (d *openstackDiscovery) clusterClient() (client *gophercloud.ServiceClient, err error) {
	if d.clients.clusterClient == nil {
		return nil, fmt.Errorf("cluster client not initialized")
	}
	return d.clients.clusterClient, nil
}
