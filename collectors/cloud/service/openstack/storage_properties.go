package openstack

import (
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
)

// getParentID returns the parent ID of a volume.
// The volume can be attached to multiple resources; retrieve the first one that has a serverID assigned.
func getParentID(volume *volumes.Volume) string {
	for _, attach := range volume.Attachments {
		if attach.ServerID != "" {
			return attach.ServerID
		}
	}

	// If no attachment is available, we attach it to the project ID
	return volume.TenantID
}
