package azure

import (
	"context"
	"errors"
	"fmt"

	"confirmate.io/collectors/cloud/internal/constants"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/lmittmann/tint"
)

// storageAtRestEncryption takes encryption properties of an armstorage.Account and converts it into our respective ontology object.
func storageAtRestEncryption(account *armstorage.Account) (enc *ontology.AtRestEncryption, err error) {
	if account == nil {
		return enc, ErrEmptyStorageAccount
	}

	if account.Properties == nil || account.Properties.Encryption.KeySource == nil {
		return enc, errors.New("keySource is empty")
	} else if util.Deref(account.Properties.Encryption.KeySource) == armstorage.KeySourceMicrosoftStorage {
		enc = &ontology.AtRestEncryption{
			Type: &ontology.AtRestEncryption_ManagedKeyEncryption{
				ManagedKeyEncryption: &ontology.ManagedKeyEncryption{
					Algorithm: constants.AES256,
					Enabled:   true,
				},
			},
		}
	} else if util.Deref(account.Properties.Encryption.KeySource) == armstorage.KeySourceMicrosoftKeyvault {
		enc = &ontology.AtRestEncryption{
			Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
				CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
					Algorithm: "", // TODO(all): TBD
					Enabled:   true,
					// TODO(oxisto): This should also include the key!
					KeyUrl: util.Deref(account.Properties.Encryption.KeyVaultProperties.KeyVaultURI),
				},
			},
		}
	}

	return enc, nil
}

// anomalyDetectionEnabled returns true if Azure Advanced Threat Protection is enabled for the database.
func (d *azureDiscovery) anomalyDetectionEnabled(server *armsql.Server, db *armsql.Database) (bool, error) {
	// initialize threat protection client
	if err := d.initThreatProtectionClient(); err != nil {
		return false, err
	}

	listPager := d.clients.threatProtectionClient.NewListByDatabasePager(resourceGroupName(util.Deref(db.ID)), *server.Name, *db.Name, &armsql.DatabaseAdvancedThreatProtectionSettingsClientListByDatabaseOptions{})
	for listPager.More() {
		pageResponse, err := listPager.NextPage(context.TODO())
		if err != nil {
			err = fmt.Errorf("%s: %v", ErrGettingNextPage, err)
			return false, err
		}

		for _, value := range pageResponse.Value {
			if *value.Properties.State == armsql.AdvancedThreatProtectionStateEnabled {
				return true, nil
			}
		}
	}
	return false, nil
}

// getActivityLogging returns the activity logging information for the storage account, blob, table and file storage including their raw information
func (d *azureDiscovery) getActivityLogging(account *armstorage.Account) (activityLoggingAccount, activityLoggingBlob, activityLoggingTable, activityLoggingFile *ontology.ActivityLogging, rawAccount, rawBlob, rawTable, rawFile string) {

	var err error

	// Get ActivityLogging for the storage account
	activityLoggingAccount, rawAccount, err = d.discoverDiagnosticSettings(util.Deref(account.ID))
	if err != nil {
		log.Error("could not discover diagnostic settings for the storage account", tint.Err(err))
	}

	// Get ActivityLogging for the blob service
	activityLoggingBlob, rawBlob, err = d.discoverDiagnosticSettings(util.Deref(account.ID) + "/blobServices/default")
	if err != nil {
		log.Error("could not discover diagnostic settings for the blob service", tint.Err(err))
	}

	// Get ActivityLogging for the table service
	activityLoggingTable, rawTable, err = d.discoverDiagnosticSettings(util.Deref(account.ID) + "/tableServices/default")
	if err != nil {
		log.Error("could not discover diagnostic settings for the table service", tint.Err(err))
	}

	// Get ActivityLogging for the file service
	activityLoggingFile, rawFile, err = d.discoverDiagnosticSettings(util.Deref(account.ID) + "/fileServices/default")
	if err != nil {
		log.Error("could not discover diagnostic settings for the file service", tint.Err(err))
	}

	return

}
