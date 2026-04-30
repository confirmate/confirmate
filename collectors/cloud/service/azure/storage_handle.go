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
	"fmt"
	"log/slog"
	"strings"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/constants"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

func (d *azureCollector) handleCosmosDB(account *armcosmos.DatabaseAccountGetResults) ([]ontology.IsResource, error) {
	var (
		atRestEnc *ontology.AtRestEncryption
		err       error
		list      []ontology.IsResource
	)

	// initialize Cosmos DB client
	if err = d.initCosmosDBClient(); err != nil {
		return nil, err
	}

	// Check if KeyVaultURI is set for Cosmos DB account By default the Cosmos DB account is encrypted by Azure managed
	// keys. Optionally, it is possible to add a second encryption layer with customer key encryption. (see
	// https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-setup-customer-managed-keys?tabs=azure-portal)
	if account.Properties.KeyVaultKeyURI != nil {
		atRestEnc = &ontology.AtRestEncryption{
			Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
				CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
					Enabled: true,
					// Algorithm: algorithm, //TODO(anatheka): How do we get the algorithm? Are we available to do it by
					// the related resources?
					KeyUrl: pointer.Deref(account.Properties.KeyVaultKeyURI),
				},
			},
		}
	} else {
		atRestEnc = &ontology.AtRestEncryption{
			Type: &ontology.AtRestEncryption_ManagedKeyEncryption{
				ManagedKeyEncryption: &ontology.ManagedKeyEncryption{
					Enabled:   true,
					Algorithm: constants.AES256,
				},
			},
		}
	}

	// Create Cosmos DB database service ontology object for the database account
	//
	// TODO(oxisto): Actually, CosmosDB is a multi-model database, but for now we just model this as a document
	// database.
	dbService := &ontology.DocumentDatabaseService{
		Id:           resourceID(account.ID),
		Name:         pointer.Deref(account.Name),
		CreationTime: creationTime(account.SystemData.CreatedAt),
		GeoLocation:  location(account.Location),
		Labels:       labels(account.Tags),
		ParentId:     resourceGroupID(account.ID),
		Raw:          collector.Raw(account),
	}

	// Add Mongo DB database service
	list = append(list, dbService)

	// Check account kind and add Mongo DB databases storages
	switch pointer.Deref(account.Kind) {
	case armcosmos.DatabaseAccountKindMongoDB:
		// Get Mongo databases
		list = append(list, d.collectMongoDBDatabases(account, atRestEnc)...)
	case armcosmos.DatabaseAccountKindGlobalDocumentDB:
		log.Info("account kind not yet implemented", slog.Any("kind", armcosmos.DatabaseAccountKindGlobalDocumentDB))
	case armcosmos.DatabaseAccountKindParse:
		log.Info("account kind not yet implemented", slog.Any("kind", armcosmos.DatabaseAccountKindParse))
	default:
		log.Warn("account kind not yet implemented", slog.Any("kind", pointer.Deref(account.Kind)))
	}

	return list, nil
}

func (d *azureCollector) handleSqlServer(server *armsql.Server) ([]ontology.IsResource, error) {
	var (
		dbList               []ontology.IsResource
		anomalyDetectionList []*ontology.AnomalyDetection
		dbService            ontology.IsResource
		list                 []ontology.IsResource
	)

	// Get SQL database storages and the corresponding anomaly detection property
	dbList, anomalyDetectionList = d.getSqlDBs(server)

	// Create SQL database service voc object for SQL server
	dbService = &ontology.RelationalDatabaseService{
		Id:           resourceID(server.ID),
		Name:         pointer.Deref(server.Name),
		CreationTime: nil,
		GeoLocation:  location(server.Location),
		Labels:       labels(server.Tags),
		ParentId:     resourceGroupID(server.ID),
		Raw:          collector.Raw(server),
		// TODO(all): HttpEndpoint
		TransportEncryption: &ontology.TransportEncryption{
			Enabled:         true,
			Enforced:        true,
			Protocol:        constants.TLS,
			ProtocolVersion: tlsVersion((*string)(server.Properties.MinimalTLSVersion)),
		},
		AnomalyDetections: anomalyDetectionList,
	}

	// Add SQL database service
	list = append(list, dbService)

	// Add SQL database storages
	list = append(list, dbList...)

	return list, nil
}

func (d *azureCollector) handleStorageAccount(account *armstorage.Account, storagesList []ontology.IsResource, activityLogging *ontology.ActivityLogging, rawActivityLogging string) (*ontology.ObjectStorageService, error) {
	var (
		storageResourceIDs []string
	)

	if account == nil {
		return nil, ErrEmptyStorageAccount
	}

	// Get all object storage IDs
	for _, storage := range storagesList {
		if strings.Contains(string(storage.GetId()), accountName(pointer.Deref(account.ID))) {
			storageResourceIDs = append(storageResourceIDs, storage.GetId())
		}
	}

	te := &ontology.TransportEncryption{
		Enforced:        pointer.Deref(account.Properties.EnableHTTPSTrafficOnly),
		Enabled:         true, // cannot be disabled
		Protocol:        constants.TLS,
		ProtocolVersion: tlsVersion((*string)(account.Properties.MinimumTLSVersion)),
	}

	storageService := &ontology.ObjectStorageService{
		Id:                  resourceID(account.ID),
		Name:                pointer.Deref(account.Name),
		StorageIds:          storageResourceIDs,
		CreationTime:        creationTime(account.Properties.CreationTime),
		GeoLocation:         location(account.Location),
		Labels:              labels(account.Tags),
		ParentId:            resourceGroupID(account.ID),
		Raw:                 collector.Raw(account, rawActivityLogging),
		TransportEncryption: te,
		HttpEndpoint: &ontology.HttpEndpoint{
			Url:                 generalizeURL(pointer.Deref(account.Properties.PrimaryEndpoints.Blob)),
			TransportEncryption: te,
		},
		ActivityLogging: activityLogging,
	}

	return storageService, nil
}

func (d *azureCollector) handleFileStorage(account *armstorage.Account, fileshare *armstorage.FileShareItem, activityLogging *ontology.ActivityLogging, rawActivityLogging string) (*ontology.FileStorage, error) {
	var (
		monitoringLogDataEnabled bool
		securityAlertsEnabled    bool
	)

	if account == nil {
		return nil, ErrEmptyStorageAccount
	}

	// It is possible that the fileshare is not empty. In that case we have to check if a mandatory field is empty, so the whole disk is empty
	if fileshare == nil || fileshare.ID == nil {
		return nil, fmt.Errorf("fileshare is nil")
	}

	// Get atRestEncryptionEnabled
	enc, err := storageAtRestEncryption(account)
	if err != nil {
		return nil, fmt.Errorf("could not get file storage properties for the atRestEncryption: %w", err)
	}

	// Get monitoringLogDataEnabled and securityAlertsEnabled
	if d.defenderProperties[DefenderStorageType] != nil {
		monitoringLogDataEnabled = d.defenderProperties[DefenderStorageType].monitoringLogDataEnabled
		securityAlertsEnabled = d.defenderProperties[DefenderStorageType].securityAlertsEnabled
	}

	return &ontology.FileStorage{
		Id:           resourceID(fileshare.ID),
		Name:         pointer.Deref(fileshare.Name),
		CreationTime: creationTime(account.Properties.CreationTime), // We only have the creation time of the storage account the file storage belongs to
		GeoLocation:  location(account.Location),                    // The location is the same as the storage account
		Labels:       labels(account.Tags),                          // The storage account labels the file storage belongs to
		ParentId:     resourceIDPointer(account.ID),                 // the storage account is our parent
		Raw:          collector.Raw(account, fileshare),
		ResourceLogging: &ontology.ResourceLogging{
			MonitoringLogDataEnabled: monitoringLogDataEnabled,
			SecurityAlertsEnabled:    securityAlertsEnabled,
		},
		ActivityLogging:  activityLogging,
		AtRestEncryption: enc,
	}, nil
}

func (d *azureCollector) handleObjectStorage(account *armstorage.Account, container *armstorage.ListContainerItem, activityLogging *ontology.ActivityLogging) (*ontology.ObjectStorage, error) {
	var (
		backups                  []*ontology.Backup
		monitoringLogDataEnabled bool
		securityAlertsEnabled    bool
	)

	if account == nil {
		return nil, ErrEmptyStorageAccount
	}

	// It is possible that the container is not empty. In that case we have to check if a mandatory field is empty, so the whole disk is empty
	if container == nil || container.ID == nil {
		return nil, fmt.Errorf("container is nil")
	}

	enc, err := storageAtRestEncryption(account)
	if err != nil {
		return nil, fmt.Errorf("could not get object storage properties for the atRestEncryption: %w", err)
	}

	if d.backupMap[DataSourceTypeStorageAccountObject] != nil && d.backupMap[DataSourceTypeStorageAccountObject].backup[pointer.Deref(account.ID)] != nil {
		backups = d.backupMap[DataSourceTypeStorageAccountObject].backup[pointer.Deref(account.ID)]
	}
	backups = backupsEmptyCheck(backups)

	if d.defenderProperties[DefenderStorageType] != nil {
		monitoringLogDataEnabled = d.defenderProperties[DefenderStorageType].monitoringLogDataEnabled
		securityAlertsEnabled = d.defenderProperties[DefenderStorageType].securityAlertsEnabled
	}

	return &ontology.ObjectStorage{
		Id:               resourceID(container.ID),
		Name:             pointer.Deref(container.Name),
		CreationTime:     creationTime(account.Properties.CreationTime), // We only have the creation time of the storage account the file storage belongs to
		GeoLocation:      location(account.Location),                    // The location is the same as the storage account
		Labels:           labels(account.Tags),                          // The storage account labels the file storage belongs to
		ParentId:         resourceIDPointer(account.ID),                 // the storage account is our parent
		Raw:              collector.Raw(account, container),
		AtRestEncryption: enc,
		Immutability: &ontology.Immutability{
			Enabled: pointer.Deref(container.Properties.HasImmutabilityPolicy),
		},
		ResourceLogging: &ontology.ResourceLogging{
			MonitoringLogDataEnabled: monitoringLogDataEnabled,
			SecurityAlertsEnabled:    securityAlertsEnabled,
		},
		ActivityLogging: activityLogging,
		Backups:         backups,
		PublicAccess:    pointer.Deref(container.Properties.PublicAccess) != armstorage.PublicAccessNone,
	}, nil
}
