package azure

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v3"
)

var (
	ErrEmptyVirtualMachine = errors.New("virtual machine is empty")
)

// Collect virtual machines
func (d *azureCollector) collectVirtualMachines() ([]ontology.IsResource, error) {
	var list []ontology.IsResource

	// initialize virtual machines client
	if err := d.initVirtualMachinesClient(); err != nil {
		return nil, err
	}

	// List all VMs
	err := listPager(d,
		d.clients.virtualMachinesClient.NewListAllPager,
		d.clients.virtualMachinesClient.NewListPager,
		func(res armcompute.VirtualMachinesClientListAllResponse) []*armcompute.VirtualMachine {
			return res.Value
		},
		func(res armcompute.VirtualMachinesClientListResponse) []*armcompute.VirtualMachine {
			return res.Value
		},
		func(vm *armcompute.VirtualMachine) error {
			r, err := d.handleVirtualMachines(vm)
			if err != nil {
				return fmt.Errorf("could not handle virtual machine: %w", err)
			}

			log.Info("Adding virtual machine", slog.String("virtual machine name", r.GetName()))

			list = append(list, r)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (d *azureCollector) collectBlockStorages() ([]ontology.IsResource, error) {
	var list []ontology.IsResource

	// initialize block storages client
	if err := d.initBlockStoragesClient(); err != nil {
		return nil, err
	}

	// List all disks
	err := listPager(d,
		d.clients.blockStorageClient.NewListPager,
		d.clients.blockStorageClient.NewListByResourceGroupPager,
		func(res armcompute.DisksClientListResponse) []*armcompute.Disk {
			return res.Value
		},
		func(res armcompute.DisksClientListByResourceGroupResponse) []*armcompute.Disk {
			return res.Value
		},
		func(disk *armcompute.Disk) error {
			blockStorage, err := d.handleBlockStorage(disk)
			if err != nil {
				return fmt.Errorf("could not handle block storage: %w", err)
			}

			log.Info("Adding block storage", slog.String("block storage name", blockStorage.GetName()))

			list = append(list, blockStorage)
			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}

// Collect functions and web apps
func (d *azureCollector) collectFunctionsWebApps() ([]ontology.IsResource, error) {
	var list []ontology.IsResource

	// initialize functions client
	if err := d.initWebAppsClient(); err != nil {
		return nil, err
	}

	// List functions
	err := listPager(d,
		d.clients.webAppsClient.NewListPager,
		d.clients.webAppsClient.NewListByResourceGroupPager,
		func(res armappservice.WebAppsClientListResponse) []*armappservice.Site {
			return res.Value
		},
		func(res armappservice.WebAppsClientListByResourceGroupResponse) []*armappservice.Site {
			return res.Value
		},
		func(site *armappservice.Site) error {
			var r ontology.IsResource

			// Get configuration for detailed properties
			config, err := d.clients.webAppsClient.GetConfiguration(context.Background(),
				util.Deref(site.Properties.ResourceGroup),
				util.Deref(site.Name),
				&armappservice.WebAppsClientGetConfigurationOptions{})
			if err != nil {
				log.Error("error getting site config",
					"resourceGroup", util.Deref(site.Properties.ResourceGroup),
					"site", util.Deref(site.Name),
					"err", err)
			}

			// Check kind of site (see https://github.com/Azure/app-service-linux-docs/blob/master/Things_You_Should_Know/kind_property.md)
			switch *site.Kind {
			case "app": // Windows Web App
				r = d.handleWebApp(site, config)
			case "app,linux": // Linux Web app
				r = d.handleWebApp(site, config)
			case "app,linux,container": // Linux Container Web App
				// TODO(all): TBD
				log.Debug("Linux Container Web App Web App currently not implemented.")
			case "hyperV": // Windows Container Web App
				// TODO(all): TBD
				log.Debug("Windows Container Web App currently not implemented.")
			case "app,container,windows": // Windows Container Web App
				// TODO(all): TBD
				log.Debug("Windows Web App currently not implemented.")
			case "app,linux,kubernetes": // Linux Web App on ARC
				// TODO(all): TBD
				log.Debug("Linux Web App on ARC currently not implemented.")
			case "app,linux,container,kubernetes": // Linux Container Web App on ARC
				// TODO(all): TBD
				log.Debug("Linux Container Web App on ARC currently not implemented.")
			case "functionapp": // Function Code App
				r = d.handleFunction(site, config)
			case "functionapp,linux": // Linux Consumption Function app
				r = d.handleFunction(site, config)
			case "functionapp,linux,container,kubernetes": // Function Container App on ARC
				// TODO(all): TBD
				log.Debug("Function Container App on ARC currently not implemented.")
			case "functionapp,linux,kubernetes": // Function Code App on ARC
				// TODO(all): TBD
				log.Debug("Function Code App on ARC currently not implemented.")
			default:
				log.Debug("kind currently not supported.", slog.String("kind", util.Deref(site.Kind)))
			}

			if r != nil {
				log.Info("Adding function", slog.String("function name", r.GetName()))
				list = append(list, r)
			}

			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}
