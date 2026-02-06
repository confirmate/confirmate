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
	log = logconfig.GetLogger().With("component", "csaf-collector")
}

type csafCollector struct {
	domain string
	ctID   string
	client *http.Client
}

type CollectorOption func(d *csafCollector)

func WithProviderDomain(domain string) CollectorOption {
	return func(d *csafCollector) {
		d.domain = domain
	}
}

func WithTargetOfEvaluationID(ctID string) CollectorOption {
	return func(a *csafCollector) {
		a.ctID = ctID
	}
}

func NewTrustedProviderCollector(opts ...CollectorOption) cloud.Collector {
	d := &csafCollector{
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

func (*csafCollector) Name() string {
	return "CSAF Trusted Provider Collector"
}

func (*csafCollector) Description() string {
	return "Collector CSAF documents from a CSAF trusted provider"
}

func (d *csafCollector) TargetOfEvaluationID() string {
	return d.ctID
}

func (d *csafCollector) List() (list []ontology.IsResource, err error) {
	log.Info("fetching CSAF documents from domain", slog.String("domain", d.domain))

	return d.collectProviders()
}
