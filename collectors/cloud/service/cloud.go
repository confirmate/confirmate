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

package cloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"time"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/logconfig"
	"confirmate.io/collectors/cloud/service/aws"
	"confirmate.io/collectors/cloud/service/azure"
	"confirmate.io/collectors/cloud/service/extra/csaf"
	"confirmate.io/collectors/cloud/service/k8s"
	"confirmate.io/collectors/cloud/service/openstack"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/service"
	"confirmate.io/core/service/collection"
	"connectrpc.com/connect"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	ProviderAWS       = "aws"
	ProviderK8S       = "k8s"
	ProviderAzure     = "azure"
	ProviderOpenstack = "openstack"
	ProviderCSAF      = "csaf"

	// CloudCollectorStart is emitted at the start of a collector run.
	CloudCollectorStart CollectorEventType = iota
	// CloudCollectorFinished is emitted at the end of a collector run.
	CloudCollectorFinished

	DefaultEvidenceStoreURL = "localhost:9092"
)

var (
	log *slog.Logger

	ErrK8sAuth       = errors.New("could not authenticate to Kubernetes")
	ErrOpenstackAuth = errors.New("could not authenticate to OpenStack")
	ErrAWSAuth       = errors.New("could not authenticate to AWS")
	ErrAzureAuth     = errors.New("could not authenticate to Azure")
)

// CloudCollectorConfig holds the configuration for the cloud collector.
type CloudCollectorConfig struct {
	// provider is the cloud service provider to use for collecting resources.
	provider string

	// targetOfEvaluationID is the target of evaluation ID for which we are gathering resources.
	targetOfEvaluationID string

	// collectorToolID is the collector tool ID which is gathering the resources.
	collectorToolID string

	// collectorInterval is the interval at which collector runs are scheduled.
	collectorInterval time.Duration

	//evStreamConfig holds the configuration for the evidence store stream.
	evStreamConfig EvidenceStoreStreamConfig
}

// EvidenceStoreStreamConfig holds the configuration for the evidence store stream.
type EvidenceStoreStreamConfig struct {
	targetAddress string
	client        *http.Client
}

// CollectorEventType defines the event types for [CollectorEvent].
type CollectorEventType int

// CollectorEvent represents an event that is emitted if certain situations happen in the collector (defined by
// [CollectorEventType]). Examples would be the start or the end of the collector. We will potentially expand this in
// the future.
type CollectorEvent struct {
	Type           CollectorEventType
	CollectorName  string
	CollectedItems int
	Time           time.Time
}

// Service is an implementation of the Clouditor Collector service (plus its experimental extensions). It should not be
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

	// scheduler is used to schedule periodic collector runs.
	scheduler *gocron.Scheduler

	// collectors is the list of collectors to use for collecting resources.
	collectors []collector.Collector

	// Events is a channel that emits collector events.
	Events chan *CollectorEvent

	// cloudConfig holds the configuration for the cloud collector.
	cloudConfig CloudCollectorConfig
}

func init() {
	log = logconfig.GetLogger()
}

// WithEvidenceStoreAddress is an option to configure the evidence store service gRPC address.
func WithEvidenceStoreAddress(target string, client *http.Client) service.Option[Service] {

	return func(s *Service) {
		log.Info("Evidence Store URL is set", slog.String("target", target))

		s.cloudConfig.evStreamConfig.targetAddress = target
		s.cloudConfig.evStreamConfig.client = client
	}
}

// WithTargetOfEvaluationID is an option to configure the target of evaluation ID for which resources will be collected.
func WithTargetOfEvaluationID(ID string) service.Option[Service] {
	return func(svc *Service) {
		log.Info("Target of Evaluation ID is set", "targetOfEvaluationID", ID)

		svc.cloudConfig.targetOfEvaluationID = ID
	}
}

// WithCollectorToolID is an option to configure the collector tool ID that is used to collect resources.
func WithCollectorToolID(ID string) service.Option[Service] {
	return func(svc *Service) {
		log.Info("Evidence Collector Tool ID is set", "collectorToolID", ID)

		svc.cloudConfig.collectorToolID = ID
	}
}

// WithProvider is an option to set the provider for collecting.
func WithProvider(provider string) service.Option[Service] {
	return func(svc *Service) {
		svc.cloudConfig.provider = provider
	}
}

// WithAdditionalCollectors is an option to add additional collectors for collecting. Note: These are added in
// addition to the one created by [WithProvider].
func WithAdditionalCollectors(collectors []collector.Collector) service.Option[Service] {
	return func(s *Service) {
		s.collectors = append(s.collectors, collectors...)
	}
}

// WithCollectorInterval is an option to set the collector interval. If not set, the collector is set to 5 minutes.
func WithCollectorInterval(interval time.Duration) service.Option[Service] {
	return func(svc *Service) {
		svc.cloudConfig.collectorInterval = interval
	}
}

func NewService(opts ...service.Option[Service]) *Service {
	var s *Service

	s = newService(opts...)

	// Set evidence store client and stream
	s.GetStream()

	return s
}

func newService(opts ...service.Option[Service]) *Service {
	var s *Service

	s = &Service{
		scheduler: gocron.NewScheduler(time.UTC),
		Events:    make(chan *CollectorEvent),
		cloudConfig: CloudCollectorConfig{
			targetOfEvaluationID: config.DefaultTargetOfEvaluationID,
			collectorToolID:      config.DefaultEvidenceCollectorToolID,
			provider:             "",
			evStreamConfig: EvidenceStoreStreamConfig{
				targetAddress: DefaultEvidenceStoreURL,
				client:        service.DefaultHTTPClient,
			},
			collectorInterval: 5 * time.Minute, // Default collector interval is 5 minutes
		},
	}

	// Apply any options
	for _, o := range opts {
		o(s)
	}

	return s
}

// buildCollectors creates the configured collectors without starting the standalone cloud scheduler.
func buildCollectors(cmd *cli.Command, opts ...service.Option[Service]) (collectors []collection.Collector, err error) {
	var (
		svc             *Service
		cloudCollectors []collector.Collector
	)

	svc = newService(opts...)
	cloudCollectors, err = svc.buildCollectors(cmd)
	if err != nil {
		return nil, err
	}

	collectors = make([]collection.Collector, 0, len(cloudCollectors))
	for _, collector := range cloudCollectors {
		collectors = append(collectors, collector)
	}

	return collectors, nil
}

func (svc *Service) Init(ctx context.Context, cmd *cli.Command) {
	var err error

	// Automatically start the collector, if we have this flag enabled
	if cmd.Bool("collector-auto-start") {
		go func() {
			err = svc.Start(cmd)
			if err != nil {
				log.Error("Could not automatically start collector", tint.Err(err))
			}
		}()
	}
}

func (svc *Service) Shutdown() {
	svc.evidenceStoreStream.CloseRequest()
	svc.scheduler.Stop()
}

func (svc *Service) buildCollectors(cmd *cli.Command) (collectors []collector.Collector, err error) {
	var (
		provider      string
		optsAzure     = []azure.CollectorOption{}
		optsOpenstack = []openstack.CollectorOption{}
	)

	collectors = append(collectors, svc.collectors...)
	provider = svc.cloudConfig.provider
	if provider == "" {
		return collectors, nil
	}

	switch {
	case provider == ProviderAzure:
		authorizer, authErr := azure.NewAuthorizer()
		if authErr != nil {
			err = fmt.Errorf("%v: %v", ErrAzureAuth, authErr)
			log.Error("authorization error", tint.Err(err))
			return nil, err
		}

		optsAzure = append(optsAzure, azure.WithAuthorizer(authorizer), azure.WithTargetOfEvaluationID(svc.cloudConfig.targetOfEvaluationID))
		if rg := cmd.String("collector-resource-group"); rg != "" {
			optsAzure = append(optsAzure, azure.WithResourceGroup(rg))
		}
		collectors = append(collectors, azure.NewAzureCollector(optsAzure...))
	case provider == ProviderK8S:
		k8sClient, authErr := k8s.AuthFromKubeConfig()
		if authErr != nil {
			err = fmt.Errorf("%v: %v", ErrK8sAuth, authErr)
			log.Error("authorization error", tint.Err(err))
			return nil, err
		}
		collectors = append(collectors,
			k8s.NewKubernetesComputeCollector(k8sClient, svc.cloudConfig.targetOfEvaluationID),
			k8s.NewKubernetesNetworkCollector(k8sClient, svc.cloudConfig.targetOfEvaluationID),
			k8s.NewKubernetesStorageCollector(k8sClient, svc.cloudConfig.targetOfEvaluationID))
	case provider == ProviderAWS:
		awsClient, authErr := aws.NewClient()
		if authErr != nil {
			err = fmt.Errorf("%v: %v", ErrAWSAuth, authErr)
			log.Error("authorization error", tint.Err(err))
			return nil, err
		}
		collectors = append(collectors,
			aws.NewAwsStorageCollector(awsClient, svc.cloudConfig.targetOfEvaluationID),
			aws.NewAwsComputeCollector(awsClient, svc.cloudConfig.targetOfEvaluationID))
	case provider == ProviderOpenstack:
		authorizer, authErr := openstack.NewAuthorizer()
		if authErr != nil {
			err = fmt.Errorf("%v: %v", ErrOpenstackAuth, authErr)
			log.Error("authorization error", tint.Err(err))
			return nil, err
		}

		optsOpenstack = append(optsOpenstack, openstack.WithAuthorizer(authorizer), openstack.WithTargetOfEvaluationID(svc.cloudConfig.targetOfEvaluationID))
		collectors = append(collectors, openstack.NewOpenstackCollector(optsOpenstack...))
	case provider == ProviderCSAF:
		var (
			domain string
			opts   []csaf.CollectorOption
		)

		domain = cmd.String("collector-csaf-domain")
		if domain != "" {
			opts = append(opts, csaf.WithProviderDomain(domain))
		}
		collectors = append(collectors, csaf.NewTrustedProviderCollector(opts...))
	default:
		err = fmt.Errorf("provider '%s' not known", provider)
		log.Error("provider not known", "provider", provider, "error", err)
		return nil, err
	}

	return collectors, nil
}

// Start collector
func (svc *Service) Start(cmd *cli.Command) (err error) {
	log.Info("Starting collector")
	svc.scheduler.TagsUnique()

	svc.collectors, err = svc.buildCollectors(cmd)
	if err != nil {
		return err
	}

	for _, v := range svc.collectors {
		log.Info("Scheduling collector", "name", v.Name(), "id", v.ID(), "interval_min", svc.cloudConfig.collectorInterval.Minutes())

		_, err = svc.scheduler.
			Every(svc.cloudConfig.collectorInterval).
			Tag(v.ID()).
			Do(svc.StartCollector, v)
		if err != nil {
			newError := fmt.Errorf("could not schedule job for {%s}: %v", v.Name(), err)
			log.Error("schedule error", "collector", v.Name(), "error", newError)
			return fmt.Errorf("%s", newError)
		}
	}

	svc.scheduler.StartAsync()

	return nil
}

func (svc *Service) StartCollector(collector collector.Collector) {
	var (
		err  error
		list []ontology.IsResource
		ev   *evidence.Evidence
	)

	go func() {
		svc.Events <- &CollectorEvent{
			Type:          CloudCollectorStart,
			CollectorName: collector.Name(),
			Time:          time.Now(),
		}
	}()

	list, err = collector.Collect()

	if err != nil {
		log.Error("Could not retrieve resources from collector", "collector", collector.Name(), tint.Err(err))
		return
	}

	// Notify event listeners that the collector is finished
	go func() {
		svc.Events <- &CollectorEvent{
			Type:           CloudCollectorFinished,
			CollectorName:  collector.Name(),
			CollectedItems: len(list),
			Time:           time.Now(),
		}
	}()

	for _, resource := range list {
		ev = &evidence.Evidence{
			Id:                   uuid.New().String(),
			TargetOfEvaluationId: svc.GetTargetOfEvaluationId(),
			Timestamp:            timestamppb.Now(),
			ToolId:               svc.cloudConfig.collectorToolID,
			Resource:             ontology.ProtoResource(resource),
		}

		// Only enabled related evidences for some specific resources for now
		if slices.Contains(ontology.ResourceTypes(resource), "SecurityAdvisoryService") {
			edges := ontology.Related(resource)
			for _, edge := range edges {
				ev.ExperimentalRelatedResourceIds = append(ev.ExperimentalRelatedResourceIds, edge.Value)
			}
		}

		err = svc.storeEvidence(&evidence.StoreEvidenceRequest{Evidence: ev})
		if err != nil {
			continue
		}
	}
}

func (svc *Service) storeEvidence(req *evidence.StoreEvidenceRequest) (err error) {
	var res *evidence.StoreEvidencesResponse

	svc.evidenceStoreStream = svc.GetStream()
	err = svc.evidenceStoreStream.Send(req)
	if err != nil {
		err = fmt.Errorf("could not send evidence to evidence store service (%s): %w", svc.cloudConfig.evStreamConfig.targetAddress, err)
		log.Error("send evidence error", "address", svc.cloudConfig.evStreamConfig.targetAddress, tint.Err(err))
		svc.checkStreamError(err)
		return err
	}

	res, err = svc.evidenceStoreStream.Receive()
	if err != nil {
		err = fmt.Errorf("could not receive evidence-store response from (%s): %w", svc.cloudConfig.evStreamConfig.targetAddress, err)
		log.Error("receive evidence-store response error", "address", svc.cloudConfig.evStreamConfig.targetAddress, tint.Err(err))
		svc.checkStreamError(err)
		return err
	}

	if res.GetStatus() != evidence.EvidenceStatus_EVIDENCE_STATUS_OK {
		err = fmt.Errorf("evidence store rejected evidence (%s): %s", svc.cloudConfig.evStreamConfig.targetAddress, res.GetStatusMessage())
		log.Error("evidence-store rejected evidence", "address", svc.cloudConfig.evStreamConfig.targetAddress, "status", res.GetStatus().String(), "message", res.GetStatusMessage())
		svc.checkStreamError(err)
		return err
	}

	return nil
}

// GetTargetOfEvaluationId implements TargetOfEvaluationRequest for this service. This is a little trick, so that we can call
// CheckAccess directly on the service. This is necessary because the collector service itself is tied to a specific
// target of evaluation ID, instead of the individual requests that are made against the service.
func (svc *Service) GetTargetOfEvaluationId() string {
	return svc.cloudConfig.targetOfEvaluationID
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
		log.Info("Re-establishing stream to Evidence Store", "address", svc.cloudConfig.evStreamConfig.targetAddress)
		svc.evidenceStoreStream = svc.evidenceStoreClient.StoreEvidences(context.Background())
		svc.dead = false
	}

	return svc.evidenceStoreStream
}

// checkStreamError checks whether the current evidence store stream can no longer be reused. If so, it marks the stream as dead.
func (svc *Service) checkStreamError(err error) {
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Info("Stream to Evidence Store closed with EOF", "address", svc.cloudConfig.evStreamConfig.targetAddress)
		} else {
			// Some other error than EOF occurred
			log.Error("Error in Evidence Store stream", "address", svc.cloudConfig.evStreamConfig.targetAddress, tint.Err(err))

			// Close the stream gracefully. We can ignore any error resulting from the close here
			if svc.evidenceStoreStream != nil {
				_ = svc.evidenceStoreStream.CloseRequest()
			}
		}
		svc.dead = true
	}
}
