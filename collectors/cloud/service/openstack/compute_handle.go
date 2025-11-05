package openstack

import (
	"context"
	"fmt"

	"confirmate.io/collectors/cloud/api/discovery"
	"confirmate.io/collectors/cloud/api/ontology"
	"confirmate.io/collectors/cloud/internal/util"

	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleServer creates a virtual machine resource based on the CSC Hub Ontology
func (d *openstackDiscovery) handleServer(server *servers.Server) (ontology.IsResource, error) {
	var (
		err         error
		bootLogging *ontology.BootLogging
	)

	// we cannot directly retrieve OS logging information
	// boot logging is logged in the console log
	consoleOutput := servers.ShowConsoleOutput(context.Background(), d.clients.computeClient, server.ID, servers.ShowConsoleOutputOpts{})
	if consoleOutput.Result.Err == nil {
		bootLogging = &ontology.BootLogging{
			Enabled: true,
		}
	} else {
		log.Errorf("Error getting boot logging: %s", consoleOutput.Err)
		// When an error occurs, we assume that boot logging is disabled.
		bootLogging = &ontology.BootLogging{
			Enabled: false,
		}
	}

	r := &ontology.VirtualMachine{
		Id:           server.ID,
		Name:         server.Name,
		CreationTime: timestamppb.New(server.Created),
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		Labels:            labels(server.Tags),
		ParentId:          util.Ref(server.TenantID),
		Raw:               discovery.Raw(server),
		MalwareProtection: &ontology.MalwareProtection{},
		BootLogging:       bootLogging,
		AutomaticUpdates:  &ontology.AutomaticUpdates{},
	}

	// Get attached block storage IDs
	for _, v := range server.AttachedVolumes {
		r.BlockStorageIds = append(r.BlockStorageIds, v.ID)
	}

	// Get attached network interface IDs
	r.NetworkInterfaceIds, err = d.getAttachedNetworkInterfaces(server.ID)
	if err != nil {
		return nil, fmt.Errorf("could not discover attached network interfaces: %w", err)
	}

	log.Infof("Adding server '%s", server.Name)

	return r, nil
}
