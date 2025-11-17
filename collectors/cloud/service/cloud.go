package cloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/service/aws"
	"confirmate.io/collectors/cloud/service/azure"
	"confirmate.io/collectors/cloud/service/extra/csaf"
	"confirmate.io/collectors/cloud/service/k8s"
	"confirmate.io/collectors/cloud/service/openstack"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/service"
	"connectrpc.com/connect"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	ProviderAWS       = "aws"
	ProviderK8S       = "k8s"
	ProviderAzure     = "azure"
	ProviderOpenstack = "openstack"
	ProviderCSAF      = "csaf"

	// CloudCollectorStart is emitted at the start of a discovery run.
	CloudCollectorStart DiscoveryEventType = iota
	// CloudCollectorFinished is emitted at the end of a discovery run.
	CloudCollectorFinished

	DefaultDiscoveryAutoStartFlag = true

	DefaultEvidenceStoreURL = "localhost:9092"
)

var (
	log *logrus.Entry

	ErrK8sAuth       = errors.New("could not authenticate to Kubernetes")
	ErrOpenstackAuth = errors.New("could not authenticate to OpenStack")
	ErrAWSAuth       = errors.New("could not authenticate to AWS")
	ErrAzureAuth     = errors.New("could not authenticate to Azure")
)

// CloudCollectorConfig holds the configuration for the cloud collector.
type CloudCollectorConfig struct {
	// Provider is the list of cloud service providers to use for discovering resources.
	Provider []string
	// TargetOfEvaluationID is the target of evaluation ID for which we are gathering resources.
	TargetOfEvaluationID string
	// CollectorToolID is the collector tool ID which is gathering the resources.
	CollectorToolID string
	//evStreamConfig holds the configuration for the evidence store stream.
	evStreamConfig EvidenceStoreStreamConfig
	// AutoStart indicates whether the discovery should be automatically started when the service is initialized.
	AutoStart     bool
	ResourceGroup string
	CSAFDomain    string
}

// EvidenceStoreStreamConfig holds the configuration for the evidence store stream.
type EvidenceStoreStreamConfig struct {
	targetAddress string
	client        *http.Client
}

// TODO(anatheka): Delete!?!
// // DefaultServiceSpec returns a [launcher.ServiceSpec] for this [Service] with all necessary options retrieved from the
// // config system.
// func DefaultServiceSpec() launcher.ServiceSpec {
// 	var providers []string

// 	// If no CSPs for discovering are given, take all implemented discoverers
// 	if len(CollectorProviderFlag) == 0 {
// 		providers = []string{ProviderAWS, ProviderAzure, ProviderK8S}
// 	} else {
// 		providers = CollectorProviderFlag
// 	}

// 	return launcher.NewServiceSpec(
// 		NewService,
// 		nil,
// 		nil,
// 		// WithOAuth2Authorizer(config.ClientCredentials()),
// 		WithTargetOfEvaluationID(TargetOfEvaluationID),
// 		WithProviders(providers),
// 		WithEvidenceStoreAddress(EvidenceStoreURL),
// 	)
// }

// DiscoveryEventType defines the event types for [DiscoveryEvent].
type DiscoveryEventType int

// DiscoveryEvent represents an event that is emitted if certain situations happen in the discoverer (defined by
// [DiscoveryEventType]). Examples would be the start or the end of the discovery. We will potentially expand this in
// the future.
type DiscoveryEvent struct {
	Type            DiscoveryEventType
	DiscovererName  string
	DiscoveredItems int
	Time            time.Time
}

// Service is an implementation of the Clouditor Discovery service (plus its experimental extensions). It should not be
// used directly, but rather the NewService constructor should be used.
type Service struct {
	// evidenceStoreClient holds the client to communicate with the evidence store service.
	evidenceStoreClient evidenceconnect.EvidenceStoreClient
	// evidenceStoreStream is the stream used to send evidences to the evidence store service.
	evidenceStoreStream *connect.BidiStreamForClient[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse]
	// dead indicates whether the stream is dead.
	dead bool
	// streamMu is used to synchronize access to the evidence store stream.
	streamMu sync.RWMutex

	// scheduler is used to schedule periodic discovery runs.
	scheduler *gocron.Scheduler

	// collectors is the list of collectors to use for discovering resources.
	collectors []cloud.Collector

	// discoveryInterval is the interval at which discovery runs are scheduled.
	discoveryInterval time.Duration

	// Events is a channel that emits discovery events.
	Events chan *DiscoveryEvent

	// TODO(anatheka): Refactor ctID and collectorID into cloudConfig
	// ctID is the target of evaluation ID for which we are gathering resources.
	ctID string

	// collectorID is the evidence collector tool ID which is gathering the resources.
	collectorID string

	// cloudConfig holds the configuration for the cloud collector.
	cloudConfig CloudCollectorConfig
}

func init() {
	log = logrus.WithField("component", "discovery")
}

// WithEvidenceStoreAddress is an option to configure the evidence store service gRPC address.
func WithEvidenceStoreAddress(target string, client *http.Client) service.Option[*Service] {

	return func(s *Service) {
		log.Infof("Evidence Store URL is set to %s", target)

		s.cloudConfig.evStreamConfig.targetAddress = target
		s.cloudConfig.evStreamConfig.client = client
	}
}

// WithTargetOfEvaluationID is an option to configure the target of evaluation ID for which resources will be discovered.
func WithTargetOfEvaluationID(ID string) service.Option[*Service] {
	return func(svc *Service) {
		log.Infof("Target of Evaluation ID is set to %s", ID)

		svc.ctID = ID
	}
}

// WithEvidenceCollectorToolID is an option to configure the collector tool ID that is used to discover resources.
func WithEvidenceCollectorToolID(ID string) service.Option[*Service] {
	return func(svc *Service) {
		log.Infof("Evidence Collector Tool ID is set to %s", ID)

		svc.collectorID = ID
	}
}

// TODO(all): Do we need that anymore?
// // WithOAuth2Authorizer is an option to use an OAuth 2.0 authorizer
// func WithOAuth2Authorizer(config *clientcredentials.Config) service.Option[*Service] {
// 	return func(svc *Service) {
// 		svc.evidenceStore.SetAuthorizer(api.NewOAuthAuthorizerFromClientCredentials(config))
// 	}
// }

// WithProviders is an option to set providers for discovering
func WithProviders(providersList []string) service.Option[*Service] {
	if len(providersList) == 0 {
		newError := errors.New("no providers given")
		log.Error(newError)
	}

	return func(svc *Service) {
		svc.cloudConfig.Provider = providersList
	}
}

// WithAdditionalDiscoverers is an option to add additional discoverers for discovering. Note: These are added in
// addition to the ones created by [WithProviders].
func WithAdditionalDiscoverers(discoverers []cloud.Collector) service.Option[*Service] {
	return func(s *Service) {
		s.collectors = append(s.collectors, discoverers...)
	}
}

// WithDiscoveryInterval is an option to set the discovery interval. If not set, the discovery is set to 5 minutes.
func WithDiscoveryInterval(interval time.Duration) service.Option[*Service] {
	return func(s *Service) {
		s.discoveryInterval = interval
	}
}

func NewService(opts ...service.Option[*Service]) *Service {
	s := &Service{
		// TODO(anatheka): Add evidence store stream
		// evidenceStoreStreams: api.NewStreamsOf(api.WithLogger[evidence.EvidenceStore_StoreEvidencesClient, *evidence.StoreEvidenceRequest](log)),
		// evidenceStore:        api.NewRPCConnection(EvidenceStoreURL), evidence.NewEvidenceStoreClient),
		scheduler:         gocron.NewScheduler(time.UTC),
		Events:            make(chan *DiscoveryEvent),
		discoveryInterval: 5 * time.Minute, // Default discovery interval is 5 minutes
		cloudConfig: CloudCollectorConfig{
			AutoStart:            DefaultDiscoveryAutoStartFlag,
			TargetOfEvaluationID: config.DefaultTargetOfEvaluationID,
			CollectorToolID:      config.DefaultEvidenceCollectorToolID,
			Provider:             []string{ProviderAWS, ProviderAzure, ProviderK8S},
			evStreamConfig: EvidenceStoreStreamConfig{
				targetAddress: DefaultEvidenceStoreURL,
				client:        http.DefaultClient,
			},
		},
	}

	// Apply any options
	for _, o := range opts {
		o(s)
	}

	// Set evidence store client and stream
	s.GetStream()

	return s
}

func (svc *Service) Init() {
	var err error

	// Automatically start the discovery, if we have this flag enabled
	if svc.cloudConfig.AutoStart {
		go func() {
			// TODO(all): Do we need that anymore?
			// <-rest.GetReadyChannel()
			err = svc.Start()
			if err != nil {
				log.Errorf("Could not automatically start discovery: %v", err)
			}
		}()
	}
}

func (svc *Service) Shutdown() {
	svc.evidenceStoreStream.CloseRequest()
	svc.scheduler.Stop()
}

// Start starts discovery
func (svc *Service) Start() (err error) {
	var (
		optsAzure     = []azure.DiscoveryOption{}
		optsOpenstack = []openstack.DiscoveryOption{}
	)

	log.Infof("Starting discovery...")
	svc.scheduler.TagsUnique()

	// Configure discoverers for given providers
	for _, provider := range svc.cloudConfig.Provider {
		switch {
		case provider == ProviderAzure:
			authorizer, err := azure.NewAuthorizer()
			if err != nil {
				err := fmt.Errorf("%v: %v", ErrAzureAuth, err)
				log.Error(err)
				return err
			}
			// Add authorizer and TargetOfEvaluationID
			optsAzure = append(optsAzure, azure.WithAuthorizer(authorizer), azure.WithTargetOfEvaluationID(svc.ctID))
			// Check if resource group is given and append to discoverer
			if svc.cloudConfig.ResourceGroup != "" {
				optsAzure = append(optsAzure, azure.WithResourceGroup(svc.cloudConfig.ResourceGroup))
			}
			svc.collectors = append(svc.collectors, azure.NewAzureDiscovery(optsAzure...))
		case provider == ProviderK8S:
			k8sClient, err := k8s.AuthFromKubeConfig()
			if err != nil {
				err := fmt.Errorf("%v: %v", ErrK8sAuth, err)
				log.Error(err)
				return err
			}
			svc.collectors = append(svc.collectors,
				k8s.NewKubernetesComputeDiscovery(k8sClient, svc.ctID),
				k8s.NewKubernetesNetworkDiscovery(k8sClient, svc.ctID),
				k8s.NewKubernetesStorageDiscovery(k8sClient, svc.ctID))
		case provider == ProviderAWS:
			awsClient, err := aws.NewClient()
			if err != nil {
				err = fmt.Errorf("%v: %v", ErrAWSAuth, err)
				log.Error(err)
				return err
			}
			svc.collectors = append(svc.collectors,
				aws.NewAwsStorageDiscovery(awsClient, svc.ctID),
				aws.NewAwsComputeDiscovery(awsClient, svc.ctID))
		case provider == ProviderOpenstack:
			authorizer, err := openstack.NewAuthorizer()
			if err != nil {
				err = fmt.Errorf("%v: %v", ErrOpenstackAuth, err)
				log.Error(err)
				return err
			}
			// Add authorizer and TargetOfEvaluationID
			optsOpenstack = append(optsOpenstack, openstack.WithAuthorizer(authorizer), openstack.WithTargetOfEvaluationID(svc.ctID))
			svc.collectors = append(svc.collectors, openstack.NewOpenstackDiscovery(optsOpenstack...))
		case provider == ProviderCSAF:
			var (
				domain string
				opts   []csaf.DiscoveryOption
			)
			domain = svc.cloudConfig.CSAFDomain
			if domain != "" {
				opts = append(opts, csaf.WithProviderDomain(domain))
			}
			svc.collectors = append(svc.collectors, csaf.NewTrustedProviderDiscovery(opts...))
		default:
			newError := fmt.Errorf("provider %s not known", provider)
			log.Error(newError)
			return fmt.Errorf("%s", newError)
		}
	}

	for _, v := range svc.collectors {
		log.Infof("Scheduling {%s} to execute every {%v} minutes...", v.Name(), svc.discoveryInterval.Minutes())

		_, err = svc.scheduler.
			Every(svc.discoveryInterval).
			Tag(v.Name()).
			Do(svc.StartDiscovery, v)
		if err != nil {
			newError := fmt.Errorf("could not schedule job for {%s}: %v", v.Name(), err)
			log.Error(newError)
			return fmt.Errorf("%s", newError)
		}
	}

	svc.scheduler.StartAsync()

	return nil
}

func (svc *Service) StartDiscovery(discoverer cloud.Collector) {
	var (
		err  error
		list []ontology.IsResource
	)

	go func() {
		svc.Events <- &DiscoveryEvent{
			Type:           CloudCollectorStart,
			DiscovererName: discoverer.Name(),
			Time:           time.Now(),
		}
	}()

	list, err = discoverer.List()

	if err != nil {
		log.Errorf("Could not retrieve resources from discoverer '%s': %v", discoverer.Name(), err)
		return
	}

	// Notify event listeners that the discoverer is finished
	go func() {
		svc.Events <- &DiscoveryEvent{
			Type:            CloudCollectorFinished,
			DiscovererName:  discoverer.Name(),
			DiscoveredItems: len(list),
			Time:            time.Now(),
		}
	}()

	for _, resource := range list {
		e := &evidence.Evidence{
			Id:                   uuid.New().String(),
			TargetOfEvaluationId: svc.GetTargetOfEvaluationId(),
			Timestamp:            timestamppb.Now(),
			ToolId:               svc.collectorID,
			Resource:             ontology.ProtoResource(resource),
		}

		// Only enabled related evidences for some specific resources for now
		if slices.Contains(ontology.ResourceTypes(resource), "SecurityAdvisoryService") {
			edges := ontology.Related(resource)
			for _, edge := range edges {
				e.ExperimentalRelatedResourceIds = append(e.ExperimentalRelatedResourceIds, edge.Value)
			}
		}

		// Get or create evidence store stream
		svc.evidenceStoreStream = svc.GetStream()
		err := svc.evidenceStoreStream.Send(&evidence.StoreEvidenceRequest{Evidence: e})
		if err != nil {
			err = fmt.Errorf("could not send evidence to evidence store service (%s): %w", svc.cloudConfig.evStreamConfig.targetAddress, err)
			log.Error(err)
			svc.checkStreamError(err)
			continue
		}
	}
}

// GetTargetOfEvaluationId implements TargetOfEvaluationRequest for this service. This is a little trick, so that we can call
// CheckAccess directly on the service. This is necessary because the discovery service itself is tied to a specific
// target of evaluation ID, instead of the individual requests that are made against the service.
func (svc *Service) GetTargetOfEvaluationId() string {
	return svc.ctID
}

// TODO(all): Maybe add a generic in core that can be used by all services to manage streams?
// GetStream returns the evidence store stream used to send evidences to the evidence store service.
func (svc *Service) GetStream() *connect.BidiStreamForClient[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse] {
	svc.streamMu.Lock()
	defer svc.streamMu.Unlock()

	if svc.evidenceStoreClient == nil {
		svc.evidenceStoreClient = evidenceconnect.NewEvidenceStoreClient(svc.cloudConfig.evStreamConfig.client, svc.cloudConfig.evStreamConfig.targetAddress)
	}

	stream := svc.evidenceStoreStream

	if stream != nil && !svc.dead {
		return stream
	} else if svc.dead || stream == nil {
		// If the stream is dead, we need to create a new one
		log.Infof("Re-establishing stream to Evidence Store (%s)...", svc.cloudConfig.evStreamConfig.targetAddress)
		svc.evidenceStoreStream = svc.evidenceStoreClient.StoreEvidences(context.Background())
		svc.dead = false
	}

	return svc.evidenceStoreStream
}

// checkStreamError checks if there was an streaming error in the evidence store stream. If there was an error, it marks the stream as dead.
func (svc *Service) checkStreamError(err error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Infof("Stream to Evidence Store (%s) closed with EOF", svc.cloudConfig.evStreamConfig.targetAddress)
		} else {
			// Some other error than EOF occurred
			log.Errorf("Error when sending message to Evidence Store (%s): %v", svc.cloudConfig.evStreamConfig.targetAddress, err)

			// Close the stream gracefully. We can ignore any error resulting from the close here
			_ = svc.evidenceStoreStream.CloseRequest()
		}
		svc.dead = true
	}
}
