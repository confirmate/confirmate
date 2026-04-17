package collector

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
)

// vmListClient is the minimal SDK surface used by this collector.
type vmListClient interface {
	NewListPager(resourceGroupName string, options *armcompute.VirtualMachinesClientListOptions) *runtime.Pager[armcompute.VirtualMachinesClientListResponse]
	NewListAllPager(options *armcompute.VirtualMachinesClientListAllOptions) *runtime.Pager[armcompute.VirtualMachinesClientListAllResponse]
}

// AzureVMFetcher fetches virtual machines from one Azure resource group.
type AzureVMFetcher struct {
	resourceGroup string
	client        vmListClient
}

// NewAzureVMFetcher creates an Azure VM fetcher using a token credential.
func NewAzureVMFetcher(subscriptionID string, credential azcore.TokenCredential, resourceGroup string, options *arm.ClientOptions) (fetcher *AzureVMFetcher, err error) {
	var client *armcompute.VirtualMachinesClient

	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}
	if credential == nil {
		return nil, fmt.Errorf("credential is required")
	}
	client, err = armcompute.NewVirtualMachinesClient(subscriptionID, credential, options)
	if err != nil {
		return nil, fmt.Errorf("creating virtual machines client: %w", err)
	}

	fetcher = &AzureVMFetcher{
		resourceGroup: resourceGroup,
		client:        client,
	}

	return fetcher, nil
}

// ListVirtualMachines retrieves VMs either from one resource group or from the whole subscription.
func (f *AzureVMFetcher) ListVirtualMachines(ctx context.Context) (vms []AzureVirtualMachine, err error) {
	if f.resourceGroup != "" {
		pager := f.client.NewListPager(f.resourceGroup, nil)

		for pager.More() {
			var page armcompute.VirtualMachinesClientListResponse

			page, err = pager.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("listing VMs for resource group %q: %w", f.resourceGroup, err)
			}

			for _, item := range page.Value {
				if item == nil {
					continue
				}

				vm := mapAzureVM(item)

				if vm.ID == "" {
					continue
				}
				if vm.Name == "" {
					vm.Name = vm.ID
				}

				vms = append(vms, vm)
			}
		}

		return vms, nil
	}

	allPager := f.client.NewListAllPager(nil)
	for allPager.More() {
		var page armcompute.VirtualMachinesClientListAllResponse

		page, err = allPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing VMs for subscription: %w", err)
		}

		for _, item := range page.Value {
			if item == nil {
				continue
			}

			vm := mapAzureVM(item)

			if vm.ID == "" {
				continue
			}
			if vm.Name == "" {
				vm.Name = vm.ID
			}

			vms = append(vms, vm)
		}
	}

	return vms, nil
}

func mapAzureVM(item *armcompute.VirtualMachine) AzureVirtualMachine {
	vm := AzureVirtualMachine{
		ID:   deref(item.ID),
		Name: deref(item.Name),
		Tags: mapTags(item.Tags),
	}

	if item.Location != nil {
		vm.Region = *item.Location
	}

	if item.Properties != nil && item.Properties.DiagnosticsProfile != nil && item.Properties.DiagnosticsProfile.BootDiagnostics != nil {
		boot := item.Properties.DiagnosticsProfile.BootDiagnostics
		vm.BootLoggingEnabled = boot.Enabled
		vm.BootLoggingStoreURI = deref(boot.StorageURI)
	}

	return vm
}

func mapTags(tags map[string]*string) map[string]string {
	if len(tags) == 0 {
		return nil
	}

	mapped := make(map[string]string, len(tags))
	for key, value := range tags {
		mapped[key] = deref(value)
	}

	return mapped
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
