package azure

import (
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
)

func getComputeStringList(values []ontology.IsResource) []string {
	var list []string

	for _, value := range values {
		list = append(list, value.GetId())

	}

	return list
}

func getInternetAccessibleEndpoint(status *armmachinelearning.PublicNetworkAccess) bool {
	// Check if status is empty
	if status == nil {
		return false
	}

	return util.Deref(status) == armmachinelearning.PublicNetworkAccessEnabled
}

// getResourceLogging returns true if application insights contains a string (applicationInsights enabled), otherwise it returns false
func getResourceLogging(log *string) *ontology.ResourceLogging {
	// Check if logging service storage is available
	if util.Deref(log) == "" {
		return &ontology.ResourceLogging{
			Enabled: false,
		}
	}

	return &ontology.ResourceLogging{
		Enabled:           true,
		LoggingServiceIds: []string{resourceID(log)},
	}
}

func getAtRestEncryption(enc *armmachinelearning.EncryptionProperty) (atRestEnc *ontology.AtRestEncryption) {

	// If the encryption property is nil, the ML workspace has managed key encryption in use
	if enc == nil {
		return &ontology.AtRestEncryption{
			Type: &ontology.AtRestEncryption_ManagedKeyEncryption{
				ManagedKeyEncryption: &ontology.ManagedKeyEncryption{
					Enabled:   true,
					Algorithm: AES256,
				},
			},
		}
	}

	if util.Deref(enc.KeyVaultProperties.KeyVaultArmID) == "" {
		atRestEnc = &ontology.AtRestEncryption{
			Type: &ontology.AtRestEncryption_ManagedKeyEncryption{
				ManagedKeyEncryption: &ontology.ManagedKeyEncryption{
					Enabled:   getEncryptionStatus(enc.Status),
					Algorithm: AES256,
				},
			},
		}
	} else {
		atRestEnc = &ontology.AtRestEncryption{
			Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
				CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
					Enabled: getEncryptionStatus(enc.Status),
					KeyUrl:  resourceID(enc.KeyVaultProperties.KeyVaultArmID),
				},
			},
		}
	}

	return atRestEnc

}

func getEncryptionStatus(enc *armmachinelearning.EncryptionStatus) bool {
	return util.Deref(enc) == armmachinelearning.EncryptionStatusEnabled
}
