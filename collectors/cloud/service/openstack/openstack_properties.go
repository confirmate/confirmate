package openstack

import (
	"context"
	"fmt"

	"confirmate.io/core/util"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/attachinterfaces"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/containerinfra/v1/clusters"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// labels converts the resource tags to the ontology label
func labels(tags *[]string) map[string]string {
	l := make(map[string]string)

	for _, tag := range util.Deref(tags) {
		l[tag] = ""
	}

	return l
}

// getAttachedNetworkInterfaces gets the attached network interfaces to the given serverID.
func (d *openstackDiscovery) getAttachedNetworkInterfaces(serverID string) ([]string, error) {
	var (
		list []string
		err  error
	)

	if err = d.authorize(); err != nil {
		return nil, fmt.Errorf("could not authorize openstack: %w", err)
	}

	err = attachinterfaces.List(d.clients.computeClient, serverID).EachPage(context.Background(), func(_ context.Context, p pagination.Page) (bool, error) {
		ifc, err := attachinterfaces.ExtractInterfaces(p)
		if err != nil {
			return false, fmt.Errorf("could not extract network interface from page: %w", err)
		}

		for _, i := range ifc {
			list = append(list, i.NetID)
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not list network interfaces: %w", err)
	}

	return list, nil
}

// setProjectInfo stores the project ID and name based on the given resource
func (d *openstackDiscovery) setProjectInfo(x interface{}) {

	switch v := x.(type) {
	case []volumes.Volume:
		d.project.projectID = v[0].TenantID
		d.project.projectName = v[0].TenantID // it is not possible to extract the project name
	case []servers.Server:
		d.project.projectID = v[0].TenantID
		d.project.projectName = v[0].TenantID // it is not possible to extract the project name
	case []networks.Network:
		d.project.projectID = v[0].TenantID
		d.project.projectName = v[0].TenantID // it is not possible to extract the project name
	case []clusters.Cluster:
		d.project.projectID = v[0].ProjectID
		d.project.projectName = v[0].ProjectID // it is not possible to extract the project name
	default:
		log.Debug("no known resource type found")
	}
}
