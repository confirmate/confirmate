// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package azure

import (
	"context"
	"errors"
	"fmt"

	"confirmate.io/collectors/cloud/internal/constants"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/lmittmann/tint"
)

// collectBackupVaults receives all backup vaults in the subscription.
// Since the backups for storage and compute are collected together, the collector is performed here and results are stored in the azureCollector receiver.
func (d *azureCollector) collectBackupVaults() error {

	if len(d.backupMap) > 0 {
		log.Debug("Backup Vaults already collected.")
		return nil
	}

	// initialize backup vaults client
	if err := d.initBackupVaultsClient(); err != nil {
		return err
	}

	// initialize backup instances client
	if err := d.initBackupInstancesClient(); err != nil {
		return err
	}

	// initialize backup policies client
	if err := d.initBackupPoliciesClient(); err != nil {
		return err
	}

	// List all backup vaults
	err := listPager(d,
		d.clients.backupVaultClient.NewGetInSubscriptionPager,
		d.clients.backupVaultClient.NewGetInResourceGroupPager,
		func(res armdataprotection.BackupVaultsClientGetInSubscriptionResponse) []*armdataprotection.BackupVaultResource {
			return res.Value
		},
		func(res armdataprotection.BackupVaultsClientGetInResourceGroupResponse) []*armdataprotection.BackupVaultResource {
			return res.Value
		},
		func(vault *armdataprotection.BackupVaultResource) error {
			instances, err := d.collectBackupInstances(resourceGroupName(pointer.Deref(vault.ID)), pointer.Deref(vault.Name))
			if err != nil {
				err := fmt.Errorf("could not collect backup instances: %v", err)
				return err
			}

			for _, instance := range instances {
				dataSourceType := pointer.Deref(instance.Properties.DataSourceInfo.DatasourceType)

				// Get retention from backup policy
				policy, err := d.clients.backupPoliciesClient.Get(context.Background(), resourceGroupName(*vault.ID), *vault.Name, backupPolicyName(*instance.Properties.PolicyInfo.PolicyID), &armdataprotection.BackupPoliciesClientGetOptions{})
				if err != nil {
					err := fmt.Errorf("could not get backup policy '%s': %w", *instance.Properties.PolicyInfo.PolicyID, err)
					log.Error("err", tint.Err(err))
					continue
				}

				// TODO(all):Maybe we should differentiate the backup retention period for different resources, e.g., disk vs blobs (Metrics)
				retention := policy.BaseBackupPolicyResource.Properties.(*armdataprotection.BackupPolicy).PolicyRules[0].(*armdataprotection.AzureRetentionRule).Lifecycles[0].DeleteAfter.(*armdataprotection.AbsoluteDeleteOption).GetDeleteOption().Duration

				// Check if map entry already exists
				_, ok := d.backupMap[dataSourceType]
				if !ok {
					d.backupMap[dataSourceType] = &backup{
						backup: make(map[string][]*ontology.Backup),
					}
				}

				// Store voc.Backup in backupMap
				d.backupMap[dataSourceType].backup[pointer.Deref(instance.Properties.DataSourceInfo.ResourceID)] = []*ontology.Backup{
					{
						Enabled:         true,
						RetentionPeriod: retentionDuration(pointer.Deref(retention)),
						StorageId:       vault.ID,
						TransportEncryption: &ontology.TransportEncryption{
							Enabled:         true,
							Enforced:        true,
							Protocol:        constants.TLS,
							ProtocolVersion: 1.2, // https://learn.microsoft.com/en-us/azure/backup/transport-layer-security#why-enable-tls-12 (Last access: 04/27/2023)
						},
					},
				}
			}
			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// collectBackupInstances retrieves the instances in a given backup vault.
// Note: It is only possible to backup a storage account with all containers in it.
func (d *azureCollector) collectBackupInstances(resourceGroup, vaultName string) ([]*armdataprotection.BackupInstanceResource, error) {
	var (
		list armdataprotection.BackupInstancesClientListResponse
		err  error
	)

	if resourceGroup == "" || vaultName == "" {
		return nil, errors.New("missing resource group and/or vault name")
	}

	// List all instances in the given backup vault
	listPager := d.clients.backupInstancesClient.NewListPager(resourceGroup, vaultName, &armdataprotection.BackupInstancesClientListOptions{})
	for listPager.More() {
		list, err = listPager.NextPage(context.TODO())
		if err != nil {
			err = fmt.Errorf("%s: %v", ErrGettingNextPage, err)
			return nil, err
		}
	}

	return list.Value, nil
}
