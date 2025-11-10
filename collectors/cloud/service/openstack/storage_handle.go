package openstack

import (
	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/api/ontology"
	"confirmate.io/collectors/cloud/internal/util"

	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleBlockStorage creates a block storage resource based on the CSC Hub Ontology
func (d *openstackDiscovery) handleBlockStorage(volume *volumes.Volume) (ontology.IsResource, error) {
	// Get Name, if exits, otherwise take the ID
	name := volume.Name
	if volume.Name == "" {
		name = volume.ID
	}

	r := &ontology.BlockStorage{
		Id:           volume.ID,
		Name:         name,
		Description:  volume.Description,
		CreationTime: timestamppb.New(volume.CreatedAt),
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		ParentId: util.Ref(getParentID(volume)),
		Labels:   map[string]string{}, // Not available
		Raw:      cloud.Raw(volume),
	}

	log.Infof("Adding block storage '%s", volume.Name)

	return r, nil
}
