package csaf

import (
	"net/http"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/core/api/ontology"

	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

func init() {
	log = logrus.WithField("component", "csaf-collector")
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
	log.Infof("Fetching CSAF documents from domain %s", d.domain)

	return d.discoverProviders()
}
