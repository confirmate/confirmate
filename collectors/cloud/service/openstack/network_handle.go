package openstack

import (
	"confirmate.io/collectors/cloud/api/discovery"
	"confirmate.io/collectors/cloud/api/ontology"
	"confirmate.io/collectors/cloud/internal/util"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleNetworkInterfaces creates a network interface resource based on the CSC Hub Ontology
func (d *openstackDiscovery) handleNetworkInterfaces(network *networks.Network) (ontology.IsResource, error) {
	r := &ontology.NetworkInterface{
		Id:           network.ID,
		Name:         network.Name,
		Description:  network.Description,
		CreationTime: timestamppb.New(network.CreatedAt),
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		Labels:   labels(util.Ref(network.Tags)),
		ParentId: util.Ref(network.ProjectID),
		Raw:      discovery.Raw(network),
	}

	log.Infof("Adding network interface '%s", network.Name)

	return r, nil
}
