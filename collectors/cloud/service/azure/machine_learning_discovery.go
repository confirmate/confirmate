package azure

import (
	"context"
	"fmt"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
)

// Discover machine learning workspace
func (d *azureDiscovery) discoverMLWorkspaces() ([]ontology.IsResource, error) {
	var list []ontology.IsResource

	// initialize machine learning client
	if err := d.initMLWorkspaceClient(); err != nil {
		return nil, err
	}

	// List all ML workspaces
	serverListPager := d.clients.mlWorkspaceClient.NewListBySubscriptionPager(&armmachinelearning.WorkspacesClientListBySubscriptionOptions{})
	for serverListPager.More() {
		pageResponse, err := serverListPager.NextPage(context.TODO())
		if err != nil {
			log.Errorf("%s: %v", ErrGettingNextPage, err)
			return list, err
		}

		// Add storage, atRestEncryption (keyVault), ...?
		for _, value := range pageResponse.Value {
			// Add ML compute resources
			compute, err := d.discoverMLCompute(resourceGroupName(util.Deref(value.ID)), value)
			if err != nil {
				return nil, fmt.Errorf("could not discover ML compute resources: %w", err)
			}

			// Get string list of compute resources for the ML workspace resource
			computeList := getComputeStringList(compute)

			list = append(list, compute...)

			// Add ML mlWorkspace
			mlWorkspace, err := d.handleMLWorkspace(value, computeList)
			if err != nil {
				return nil, fmt.Errorf("could not handle ML workspace: %w", err)
			}

			log.Infof("Adding ML workspace '%s'", mlWorkspace.GetName())

			list = append(list, mlWorkspace)
		}
	}

	return list, nil
}

// discoverMLCompute discovers machine learning compute nodes
func (d *azureDiscovery) discoverMLCompute(rg string, workspace *armmachinelearning.Workspace) ([]ontology.IsResource, error) {
	var list []ontology.IsResource

	// initialize machine learning compute client
	if err := d.initMachineLearningComputeClient(); err != nil {
		return nil, err
	}

	// List all computes nodes in specific ML workspace
	serverListPager := d.clients.mlComputeClient.NewListPager(rg, util.Deref(workspace.Name), &armmachinelearning.ComputeClientListOptions{})
	for serverListPager.More() {
		pageResponse, err := serverListPager.NextPage(context.TODO())
		if err != nil {
			log.Errorf("%s: %v", ErrGettingNextPage, err)
			return list, err
		}

		for _, value := range pageResponse.Value {
			compute, err := d.handleMLCompute(value, workspace.ID)
			if err != nil {
				return nil, fmt.Errorf("could not handle ML workspace: %w", err)
			}

			if compute == nil {
				continue
			}

			log.Infof("Adding ML compute resource '%s'", compute.GetName())

			list = append(list, compute)

		}
	}

	return list, nil
}
