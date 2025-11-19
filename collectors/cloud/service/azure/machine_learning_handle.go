package azure

import (
	"log/slog"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (d *azureDiscovery) handleMLWorkspace(value *armmachinelearning.Workspace, computeList []string) (ontology.IsResource, error) {
	ml := &ontology.MachineLearningService{
		Id:                         resourceID(value.ID),
		Name:                       util.Deref(value.Name),
		CreationTime:               creationTime(value.SystemData.CreatedAt),
		GeoLocation:                location(value.Location),
		Labels:                     labels(value.Tags),
		ParentId:                   resourceGroupID(resourceIDPointer(value.ID)),
		Raw:                        cloud.Raw(value),
		InternetAccessibleEndpoint: getInternetAccessibleEndpoint(value.Properties.PublicNetworkAccess),
		StorageIds:                 []string{util.Deref(value.Properties.StorageAccount)},
		ComputeIds:                 computeList,
		Loggings: []*ontology.Logging{
			{
				Type: &ontology.Logging_ResourceLogging{
					ResourceLogging: getResourceLogging(value.Properties.ApplicationInsights),
				},
			},
		},
	}

	return ml, nil
}

// TODO(all): Should we move that to the compute file
func (d *azureDiscovery) handleMLCompute(value *armmachinelearning.ComputeResource, workspaceID *string) (ontology.IsResource, error) {
	var (
		compute   *ontology.VirtualMachine
		container *ontology.Container
		time      = &timestamppb.Timestamp{}
	)

	// Get properties vom ComputeResource
	if value.SystemData != nil && value.SystemData.CreatedAt != nil {
		time = creationTime(value.SystemData.CreatedAt)
	}

	// Get compute type specific properties for "VirtualMachine" or "ComputeInstance"
	switch c := value.Properties.(type) {
	case *armmachinelearning.ComputeInstance:
		container = &ontology.Container{
			Id:                  resourceID(value.ID),
			Name:                util.Deref(value.Name),
			CreationTime:        time,
			GeoLocation:         location(value.Location),
			Labels:              labels(value.Tags),
			ParentId:            resourceIDPointer(workspaceID),
			Raw:                 cloud.Raw(value, c.ComputeLocation),
			NetworkInterfaceIds: []string{},
		}
		return container, nil
	case *armmachinelearning.VirtualMachine:

		compute = &ontology.VirtualMachine{
			Id:                  resourceID(value.ID),
			Name:                util.Deref(value.Name),
			CreationTime:        time,
			GeoLocation:         location(value.Location),
			Labels:              labels(value.Tags),
			ParentId:            resourceIDPointer(workspaceID),
			Raw:                 cloud.Raw(value, c.ComputeLocation),
			NetworkInterfaceIds: []string{},
			BlockStorageIds:     []string{},
			MalwareProtection:   &ontology.MalwareProtection{},
		}

		return compute, nil
	}

	slog.Debug("unsupported compute resource type",
		"name", util.Deref(value.Name),
		"type", util.Deref(value.Type),
	)
	return nil, nil
}
