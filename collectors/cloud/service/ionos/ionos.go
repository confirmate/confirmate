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

// package ionos contains a Confirmate collector for IONOS Cloud environments.
package ionos

import (
	"fmt"
	"log/slog"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/logconfig"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	"github.com/google/uuid"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
	"github.com/lmittmann/tint"
)

var log *slog.Logger

func (*ionosCollector) Name() string {
	return "IONOS Cloud"
}

func (*ionosCollector) Description() string {
	return "Collector IONOS Cloud."
}

type CollectorOption func(d *ionosCollector)

// WithAuthorizer is an option to set the IONOS Cloud configuration used for authentication.
func WithAuthorizer(cfg *ionoscloud.Configuration) CollectorOption {
	return func(d *ionosCollector) {
		d.authConfig = cfg
	}
}

func WithTargetOfEvaluationID(ctID string) CollectorOption {
	return func(d *ionosCollector) {
		d.ctID = ctID
	}
}

func init() {
	log = logconfig.GetLogger().With("component", "ionos-collector")
}

type ionosCollector struct {
	ctID       string
	id         string
	authConfig *ionoscloud.Configuration // authConfig contains the IONOS Cloud configuration, which is used to authenticate against the IONOS Cloud API
	client     *ionoscloud.APIClient
}

func NewIonosCollector(opts ...CollectorOption) collector.Collector {
	d := &ionosCollector{
		ctID: config.DefaultTargetOfEvaluationID,
	}

	// Apply options
	for _, opt := range opts {
		opt(d)
	}

	seed := "ionos::" + d.ctID
	d.id = uuid.NewSHA1(uuid.NameSpaceOID, []byte(seed)).String()

	// WithAuthorizer is mandatory, since it cannot be checked directly whether WithAuthorizer was passed, we check if authConfig is set before returning the collector
	if d.authConfig == nil {
		return nil
	}

	return d
}

func (d *ionosCollector) authorize() {
	if d.client == nil {
		d.client = ionoscloud.NewAPIClient(d.authConfig)
	}
}

// NewAuthorizer returns the IONOS Cloud configuration
func NewAuthorizer() (*ionoscloud.Configuration, error) {
	cfg := ionoscloud.NewConfigurationFromEnv()
	return cfg, nil
}

// List discovers the following IONOS Cloud resource types:
// * Datacenters
// * Block storages
// * Virtual machines and network interfaces
// * Load balancers
//
// For IONOS Cloud, the datacenters have to be discovered first, since all other resources are scoped to a datacenter.
func (d *ionosCollector) List() (list []ontology.IsResource, err error) {
	d.authorize()

	log.Debug("Discover IONOS Cloud datacenters")
	dc, dcResources, err := d.collectDatacenters()
	if err != nil {
		return nil, fmt.Errorf("could not collect datacenters: %w", err)
	}
	list = append(list, dcResources...)

	for _, datacenter := range pointer.Deref(dc.Items) {
		// Collect block storage
		storage, err := d.collectBlockStorages(datacenter)
		if err != nil {
			log.Error("could not collect block storage", tint.Err(err))
		}
		list = append(list, storage...)

		// Collect virtual machines and network interfaces. Network interfaces cannot be discovered on their
		// own, only via the servers they are attached to.
		servers, err := d.collectServers(datacenter)
		if err != nil {
			log.Error("could not collect servers and network interfaces", tint.Err(err))
		}
		list = append(list, servers...)

		// Collect load balancers
		loadBalancer, err := d.collectLoadBalancers(datacenter)
		if err != nil {
			log.Error("could not collect load balancers", tint.Err(err))
		}
		list = append(list, loadBalancer...)
	}

	return list, nil
}

// Collect is the core collection contract and delegates to the existing List implementation.
func (d *ionosCollector) Collect() (list []ontology.IsResource, err error) {
	return d.List()
}

// ID returns a stable collector ID derived from collector type and target of evaluation.
func (d *ionosCollector) ID() string {
	return d.id
}

func (d *ionosCollector) TargetOfEvaluationID() string {
	return d.ctID
}
