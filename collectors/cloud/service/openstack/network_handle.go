package openstack

import (
	"log/slog"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleNetworkInterfaces creates a network interface resource based on the CSC Hub Ontology
func (d *openstackCollector) handleNetworkInterfaces(network *networks.Network) (ontology.IsResource, error) {
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
		Raw:      collector.Raw(network),
	}

	log.Info("Adding network interface", slog.String("name", network.Name))

	return r, nil
}
