package testdata

const (
	MockGRPCTarget = "localhost"

	// Azure
	MockLocationWestEurope     = "West Europe"
	MockLocationEastUs         = "eastus"
	MockSubscriptionID         = "00000000-0000-0000-0000-000000000000"
	MockSubscriptionResourceID = "/subscriptions/00000000-0000-0000-0000-000000000000"
	MockResourceGroupID        = "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/res1"
	MockResourceGroup          = "TestResourceGroup"

	// OpenStack
	// auth options
	MockOpenstackIdentityEndpoint = "https://identityHost:portNumber/v2.0" //"https://openstack.test:8888/v2.0"
	MockOpenstackUsername         = "username"
	MockOpenstackPassword         = "password"
	MockOpenstackTenantName       = "openstackTenant"

	// domain
	MockOpenstackDomainID1          = "00000000000000000000000000000000"
	MockOpenstackDomainName1        = "Domain 1"
	MockOpenstackDomainDescription1 = "This is a mock domain description (1)"

	// project
	MockOpenstackProjectID1          = "00000000000000000000000000000001"
	MockOpenstackProjectName1        = "Project 1"
	MockOpenstackProjectDescription1 = "This is a mock project description (1)."
	MockOpenstackProjectParentID1    = MockOpenstackDomainID1

	// server
	MockOpenstackServerID1      = "00000000000000000000000000000002"
	MockOpenstackServerName1    = "Server 1"
	MockOpenstackServerTenantID = MockOpenstackDomainID1

	// volume
	MockOpenstackVolumeID1      = "00000000000000000000000000000003"
	MockOpenstackVolumeName1    = "Volume 1"
	MockOpenstackVolumeTenantID = MockOpenstackDomainID1

	// network
	MockOpenstackNetworkID1   = "00000000000000000000000000000004"
	MockOpenstackNetworkName1 = "Network 1"

	// Audit Scope
	MockAuditScopeID1   = "11111111-1111-1111-1111-111111111123"
	MockAuditScopeName1 = "Mock Audit Scope 1"
	MockAuditScopeID2   = "11111111-1111-1111-1111-111111111124"
	MockAuditScopeName2 = "Mock Audit Scope 2"
	MockAuditScopeID3   = "11111111-1111-1111-1111-111111111125"
	MockAuditScopeName3 = "Mock Audit Scope 3"

	// Auth
	MockAuthUser     = "clouditor"
	MockAuthPassword = "clouditor"

	MockAuthClientID     = "client"
	MockAuthClientSecret = "secret"

	// Target of Evaluation
	MockTargetOfEvaluationID1          = "11111111-1111-1111-1111-111111111111"
	MockTargetOfEvaluationName1        = "Mock Target of Evaluation"
	MockTargetOfEvaluationDescription1 = "This is a mock target of evaluation"
	MockTargetOfEvaluationID2          = "22222222-2222-2222-2222-222222222222"
	MockTargetOfEvaluationName2        = "Another Mock Target of Evaluation"
	MockTargetOfEvaluationDescription2 = "This is another mock target of evaluation"

	// Evidence
	MockEvidenceID1     = "11111111-1111-1111-1111-111111111111"
	MockEvidenceToolID1 = "39d85e98-c3da-11ed-afa1-0242ac120002"
	MockEvidenceID2     = "22222222-2222-2222-2222-222222222222"
	MockEvidenceToolID2 = "49d85e98-c3da-11ed-afa1-0242ac120002"

	// Virtual Machine
	MockVirtualMachineID1          = "my-vm-id"
	MockVirtualMachineName1        = "my-vm-name"
	MockVirtualMachineDescription1 = "This is a mock virtual machine"
	MockVirtualMachineID2          = "my-other-vm-id"
	MockVirtualMachineName2        = "my-other-vm-name"
	MockVirtualMachineDescription2 = "This is another mock virtual machine"

	// Block Storage
	MockBlockStorageID1          = "my-block-storage-id"
	MockBlockStorageName1        = "my-block-storage-name"
	MockBlockStorageDescription1 = "This is a mock block storage"
	MockBlockStorageID2          = "my-other-block-storage-id"
	MockBlockStorageName2        = "my-other-block-storage-name"
	MockBlockStorageDescription2 = "This is another mock block storage"
)

var (
	// Resource Types
	MockVirtualMachineTypes = []string{"VirtualMachine", "Compute", "Infrastructure", "Resource"}
	MockBlockStorageTypes   = []string{"BlockStorage", "Storage", "Infrastructure", "Resource"}
)
