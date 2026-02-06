package openstack

import (
	"log/slog"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/gophercloud/gophercloud/v2/openstack/containerinfra/v1/clusters"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleCluster creates a container resource based on the CSC Hub Ontology
func (d *openstackCollector) handleCluster(cluster *clusters.Cluster) (ontology.IsResource, error) {
	r := &ontology.ContainerOrchestration{
		Id:           cluster.UUID,
		Name:         cluster.Name,
		CreationTime: timestamppb.New(cluster.CreatedAt),
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		Labels:   cluster.Labels,
		ParentId: util.Ref(cluster.ProjectID),
		Raw:      cloud.Raw(cluster),
	}

	log.Info("Adding cluster", slog.String("name", cluster.Name))

	return r, nil
}
