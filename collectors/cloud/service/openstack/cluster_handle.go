package openstack

import (
	"confirmate.io/collectors/cloud/api/discovery"
	"confirmate.io/collectors/cloud/api/ontology"
	"confirmate.io/collectors/cloud/internal/util"

	"github.com/gophercloud/gophercloud/v2/openstack/containerinfra/v1/clusters"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleCluster creates a container resource based on the CSC Hub Ontology
func (d *openstackDiscovery) handleCluster(cluster *clusters.Cluster) (ontology.IsResource, error) {
	r := &ontology.ContainerOrchestration{
		Id:           cluster.UUID,
		Name:         cluster.Name,
		CreationTime: timestamppb.New(cluster.CreatedAt),
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		Labels:   cluster.Labels,
		ParentId: util.Ref(cluster.ProjectID),
		Raw:      discovery.Raw(cluster),
	}

	log.Infof("Adding cluster '%s", cluster.Name)

	return r, nil
}
