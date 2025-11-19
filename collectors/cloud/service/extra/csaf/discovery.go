package csaf

import (
	"log/slog"
	"net/http"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/logconfig"
	"confirmate.io/core/api/ontology"
)

var log *slog.Logger

func init() {
	log = logconfig.GetLogger().With("component", "csaf-discovery")
}

type csafDiscovery struct {
	domain string
	ctID   string
	client *http.Client
}

type DiscoveryOption func(d *csafDiscovery)

func WithProviderDomain(domain string) DiscoveryOption {
	return func(d *csafDiscovery) {
		d.domain = domain
	}
}

func WithTargetOfEvaluationID(ctID string) DiscoveryOption {
	return func(a *csafDiscovery) {
		a.ctID = ctID
	}
}

func NewTrustedProviderDiscovery(opts ...DiscoveryOption) cloud.Collector {
	d := &csafDiscovery{
		ctID:   config.DefaultTargetOfEvaluationID,
		domain: "confirmate.io",
		client: http.DefaultClient,
	}

	// Apply options
	for _, opt := range opts {
		opt(d)
	}

	return d
}

func (*csafDiscovery) Name() string {
	return "CSAF Trusted Provider Discovery"
}

func (*csafDiscovery) Description() string {
	return "Discovery CSAF documents from a CSAF trusted provider"
}

func (d *csafDiscovery) TargetOfEvaluationID() string {
	return d.ctID
}

func (d *csafDiscovery) List() (list []ontology.IsResource, err error) {
	log.Info("fetching CSAF documents from domain", slog.String("domain", d.domain))

	return d.discoverProviders()
}
