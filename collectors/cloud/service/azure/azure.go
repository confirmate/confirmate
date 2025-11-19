package azure

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/logconfig"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/lmittmann/tint"
)

const (
	DefenderStorageType        = "StorageAccounts"
	DefenderVirtualMachineType = "VirtualMachines"

	DataSourceTypeDisc                 = "Microsoft.Compute/disks"
	DataSourceTypeStorageAccountObject = "Microsoft.Storage/storageAccounts/blobServices"

	Duration30Days = time.Duration(30 * time.Hour * 24)
	Duration7Days  = time.Duration(7 * time.Hour * 24)

	AES256 = "AES256"

	RetentionPeriod90Days = 90 * time.Hour * 24
)

var (
	log *slog.Logger

	ErrCouldNotAuthenticate     = errors.New("could not authenticate to Azure")
	ErrCouldNotGetSubscriptions = errors.New("could not get azure subscription")
	ErrGettingNextPage          = errors.New("could not get next page")
	ErrNoCredentialsConfigured  = errors.New("no credentials were configured")
	ErrSubscriptionNotFound     = errors.New("SubscriptionNotFound")
	ErrVaultInstanceIsEmpty     = errors.New("vault and/or instance is nil")
)

func (*azureCollector) Name() string {
	return "Azure"
}

func (*azureCollector) Description() string {
	return "Collector Azure."
}

type CollectorOption func(d *azureCollector)

func WithSender(sender policy.Transporter) CollectorOption {
	return func(d *azureCollector) {
		d.clientOptions.Transport = sender
	}
}

func WithAuthorizer(authorizer azcore.TokenCredential) CollectorOption {
	return func(d *azureCollector) {
		d.cred = authorizer
	}
}

func WithTargetOfEvaluationID(ctID string) CollectorOption {
	return func(a *azureCollector) {
		a.ctID = ctID
	}
}

// WithResourceGroup is a [CollectorOption] that scopes the collector to a specific resource group.
func WithResourceGroup(rg string) CollectorOption {
	return func(d *azureCollector) {
		d.rg = &rg
	}
}

func init() {
	log = logconfig.GetLogger().With("component", "azure-collector")
}

type azureCollector struct {
	isAuthorized bool

	sub  *armsubscription.Subscription
	cred azcore.TokenCredential
	// rg optionally contains the name of a resource group. If this is not nil, all collector calls will be scoped to the particular resource group.
	rg                 *string
	clientOptions      arm.ClientOptions
	clients            clients
	ctID               string
	backupMap          map[string]*backup
	defenderProperties map[string]*defenderProperties
}

type defenderProperties struct {
	monitoringLogDataEnabled bool
	securityAlertsEnabled    bool
}

type backup struct {
	// backup is a list of all ontology.Backup objects
	backup map[string][]*ontology.Backup
	// backupStorages is a list of all backed up ontology.Storage (ObjectStorage, BlockStorage) objects
	backupStorages []ontology.IsResource
}

type clients struct {
	// Storage
	blobContainerClient      *armstorage.BlobContainersClient
	fileStorageClient        *armstorage.FileSharesClient
	accountsClient           *armstorage.AccountsClient
	diagnosticSettingsClient *armmonitor.DiagnosticSettingsClient

	// DB
	databasesClient        *armsql.DatabasesClient
	sqlServersClient       *armsql.ServersClient
	threatProtectionClient *armsql.DatabaseAdvancedThreatProtectionSettingsClient
	cosmosDBClient         *armcosmos.DatabaseAccountsClient
	mongoDBResourcesClient *armcosmos.MongoDBResourcesClient

	// Network
	networkInterfacesClient     *armnetwork.InterfacesClient
	loadBalancerClient          *armnetwork.LoadBalancersClient
	applicationGatewayClient    *armnetwork.ApplicationGatewaysClient
	networkSecurityGroupsClient *armnetwork.SecurityGroupsClient

	// AppService
	webAppsClient *armappservice.WebAppsClient

	// Compute
	virtualMachinesClient *armcompute.VirtualMachinesClient
	blockStorageClient    *armcompute.DisksClient
	diskEncSetClient      *armcompute.DiskEncryptionSetsClient

	// Security
	defenderClient *armsecurity.PricingsClient

	// Machine Learning
	mlWorkspaceClient *armmachinelearning.WorkspacesClient
	mlComputeClient   *armmachinelearning.ComputeClient

	// Data protection
	backupPoliciesClient  *armdataprotection.BackupPoliciesClient
	backupVaultClient     *armdataprotection.BackupVaultsClient
	backupInstancesClient *armdataprotection.BackupInstancesClient

	// Resource groups
	rgClient *armresources.ResourceGroupsClient
}

func NewAzureCollector(opts ...CollectorOption) cloud.Collector {
	d := &azureCollector{
		ctID:               config.DefaultTargetOfEvaluationID,
		backupMap:          make(map[string]*backup),
		defenderProperties: make(map[string]*defenderProperties),
	}

	// Apply options
	for _, opt := range opts {
		opt(d)
	}

	return d
}

// List collects the following Azure resources types:
// - ResourceGroup resource
// - Storage resource
// - Compute resource
// - Network resource
func (d *azureCollector) List() (list []ontology.IsResource, err error) {
	if err = d.authorize(); err != nil {
		return nil, fmt.Errorf("%s: %w", ErrCouldNotAuthenticate, err)
	}

	// Collect resource group resources
	log.Info("Collect Azure resource group resources...")
	rg, err := d.collectResourceGroups()
	if err != nil {
		return nil, fmt.Errorf("could not collect resource groups: %w", err)
	}
	list = append(list, rg...)

	// Collect storage resources
	log.Info("Collect Azure storage resources...")

	// Collect Defender for X properties to add it to the required resource properties
	d.defenderProperties, err = d.collectDefender()
	if err != nil {
		return nil, fmt.Errorf("could not collect Defender for X: %w", err)
	}

	// Collect storage accounts
	storageAccounts, err := d.collectStorageAccounts()
	if err != nil {
		return nil, fmt.Errorf("could not collect storage accounts: %w", err)
	}
	list = append(list, storageAccounts...)

	// Collect sql databases
	dbs, err := d.collectSqlServers()
	if err != nil {
		return nil, fmt.Errorf("could not collect sql databases: %w", err)
	}
	list = append(list, dbs...)

	// Collect Cosmos DB
	cosmosDB, err := d.collectCosmosDB()
	if err != nil {
		return nil, fmt.Errorf("could not collect cosmos db accounts: %w", err)
	}
	list = append(list, cosmosDB...)

	// Collect compute resources
	log.Info("Collect Azure compute resources...")

	// Collect backup vaults
	err = d.collectBackupVaults()
	if err != nil {
		log.Error("could not collect backup vaults", tint.Err(err))
	}

	// Collect block storage
	storage, err := d.collectBlockStorages()
	if err != nil {
		return nil, fmt.Errorf("could not collect block storage: %w", err)
	}
	list = append(list, storage...)

	// Add backup block storages
	if d.backupMap[DataSourceTypeDisc] != nil && d.backupMap[DataSourceTypeDisc].backupStorages != nil {
		list = append(list, d.backupMap[DataSourceTypeDisc].backupStorages...)
	}

	// Collect virtual machines
	virtualMachines, err := d.collectVirtualMachines()
	if err != nil {
		return nil, fmt.Errorf("could not collect virtual machines: %w", err)
	}
	list = append(list, virtualMachines...)

	// Collect functions and web apps
	resources, err := d.collectFunctionsWebApps()
	if err != nil {
		return nil, fmt.Errorf("could not collect functions: %w", err)
	}

	list = append(list, resources...)

	// Collect network resources
	log.Info("Collect Azure network resources...")

	// Collect network interfaces
	networkInterfaces, err := d.collectNetworkInterfaces()
	if err != nil {
		return nil, fmt.Errorf("could not collect network interfaces: %w", err)
	}
	list = append(list, networkInterfaces...)

	// Collect Load Balancer
	loadBalancer, err := d.collectLoadBalancer()
	if err != nil {
		return list, fmt.Errorf("could not collect load balancer: %w", err)
	}
	list = append(list, loadBalancer...)

	// Collect Application Gateway
	ag, err := d.collectApplicationGateway()
	if err != nil {
		return list, fmt.Errorf("could not collect application gateways: %w", err)
	}
	list = append(list, ag...)

	// Collect machine learning workspaces
	mlWorkspaces, err := d.collectMLWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("could not collect machine learning workspaces: %w", err)
	}
	list = append(list, mlWorkspaces...)

	return list, nil
}

func (d *azureCollector) TargetOfEvaluationID() string {
	return d.ctID
}

func (d *azureCollector) authorize() (err error) {
	if d.isAuthorized {
		return
	}

	if d.cred == nil {
		return ErrNoCredentialsConfigured
	}

	// Create new subscriptions client
	subClient, err := armsubscription.NewSubscriptionsClient(d.cred, &d.clientOptions)
	if err != nil {
		err = fmt.Errorf("could not get new subscription client: %w", err)
		return err
	}

	// Get subscriptions
	subPager := subClient.NewListPager(nil)
	subList := make([]*armsubscription.Subscription, 0)
	for subPager.More() {
		pageResponse, err := subPager.NextPage(context.TODO())
		if err != nil {
			err = fmt.Errorf("%s: %w", ErrCouldNotGetSubscriptions, err)
			log.Error("error", tint.Err(err))
			return err
		}
		subList = append(subList, pageResponse.ListResult.Value...)
	}

	// check if list of subscriptions is empty
	if len(subList) == 0 {
		err = errors.New("list of subscriptions is empty")
		return
	}

	// get first subscription
	d.sub = subList[0]

	log.Info("Azure collector uses subscription", "subscriptionID", util.Deref(d.sub.SubscriptionID))

	d.isAuthorized = true

	return nil
}

// NewAuthorizer returns the Azure credential using one of the following authentication types (in the following order):
//
//	EnvironmentCredential
//	ManagedIdentityCredential
//	AzureCLICredential
func NewAuthorizer() (*azidentity.DefaultAzureCredential, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(ErrCouldNotAuthenticate.Error(), tint.Err(err))
		return nil, fmt.Errorf("%s: %w", ErrCouldNotAuthenticate, err)
	}

	return cred, nil
}

// collectDefender collects Defender for X services and returns a map with the following properties for each defender type
// * monitoringLogDataEnabled
// * securityAlertsEnabled
// The property will be set to the individual resources, e.g., compute, storage in the corresponding collectors
func (d *azureCollector) collectDefender() (map[string]*defenderProperties, error) {
	var pricings = make(map[string]*defenderProperties)

	// initialize defender client
	if err := d.initDefenderClient(); err != nil {
		return nil, err
	}

	// List all pricings to get the enabled Defender for X
	pricingsList, err := d.clients.defenderClient.List(context.Background(), *d.sub.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("could not collect pricings")
	}

	for _, elem := range pricingsList.Value {
		if util.Deref(elem.Properties.PricingTier) == armsecurity.PricingTierFree {
			pricings[util.Deref(elem.Name)] = &defenderProperties{
				monitoringLogDataEnabled: false,
				securityAlertsEnabled:    false,
			}
		} else {
			pricings[util.Deref(elem.Name)] = &defenderProperties{
				monitoringLogDataEnabled: true,
				securityAlertsEnabled:    true,
			}
		}
	}

	return pricings, nil
}

// listPager loops all values from a [runtime.Pager] object from the Azure SDK and issues a callback for each item. It
// takes the following arguments:
//   - d, an [azureCollector] struct,
//   - newListAllPager, a function that supplies a [runtime.Pager] listing all resources of a specific Azure client,
//   - newListByResourceGroupPager, a function that supplies a [runtime.Pager] listing all resources of a specific resource group,
//   - resToValues1, a function that takes the response from a single page of newListAllPager and returns its values,
//   - resToValues2, a function that takes the response from a single page of newListByResourceGroupPager and returns its values,
//   - callback, a function that is called for each item in every page.
//
// This function will then decide to use newListAllPager or newListByResourceGroupPager depending on whether a resource
// group scope is set in the [azureCollector] object.
//
// This function makes heavy use of the following type constraints (generics):
//   - O1, a type that represents an option argument to the newListAllPager function, e.g. *[armcompute.VirtualMachinesClientListAllOptions],
//   - R1, a type that represents the return type of the newListAllPager function, e.g. [armcompute.VirtualMachinesClientListAllResponse],
//   - O2, a type that represents an option argument to the newListByResourceGroupPager function, e.g. *[armcompute.VirtualMachinesClientListOptions],
//   - R1, a type that represents the return type of the newListAllPager function, e.g. [armcompute.VirtualMachinesClientListResponse],
//   - T, a type that represents the final resource that is supplied to the callback, e.g. *[armcompute.VirtualMachine].
func listPager[O1 any, R1 any, O2 any, R2 any, T any](
	d *azureCollector,
	newListAllPager func(options O1) *runtime.Pager[R1],
	newListByResourceGroupPager func(resourceGroupName string, options O2) *runtime.Pager[R2],
	allPagerResponseToValues func(res R1) []*T,
	allByResourceGroupPagerResponseToValues func(res R2) []*T,
	callback func(disk *T) error,
) error {
	// If the resource group is empty, we invoke the all-pager
	if d.rg == nil {
		pager := newListAllPager(*new(O1))
		// Invoke a callback for each page
		return allPages(pager, func(page R1) error {
			// Retrieve all resources of every page
			values := allPagerResponseToValues(page)
			for _, resource := range values {
				// Invoke the outer callback for each resource
				err := callback(resource)
				// We abort with the supplied error, if the callback issued an error
				if err != nil {
					return err
				}
			}

			return nil
		})
	} else {
		// Otherwise, we ivnoke the by-resource-group-pager
		pager := newListByResourceGroupPager(*d.rg, *new(O2))
		// Invoke a callback for each page
		return allPages(pager, func(page R2) error {
			// Retrieve all resources of every page
			values := allByResourceGroupPagerResponseToValues(page)
			for _, resource := range values {
				// Invoke the outer callback for each resource
				err := callback(resource)
				// We abort with the supplied error, if the callback issued an error
				if err != nil {
					return err
				}
			}

			return nil
		})
	}
}

// allPages loops through all pages of a [runtime.Pager] and issues a callback to each page.
func allPages[T any](pager *runtime.Pager[T], callback func(page T) error) error {
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("%s: %w", ErrGettingNextPage, err)
		}

		err = callback(page)
		if err != nil {
			return err
		}
	}

	return nil
}
