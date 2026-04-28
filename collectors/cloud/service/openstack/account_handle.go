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

package openstack

import (
	"log/slog"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
)

// handleDomain returns a [ontology.Account] out of an existing [domains.Domain].
func (d *openstackCollector) handleDomain(domain *domains.Domain) (ontology.IsResource, error) {
	r := &ontology.Account{
		Id:           domain.ID,
		Name:         domain.Name,
		Description:  domain.Description,
		CreationTime: nil, // domain does not have a creation date
		GeoLocation:  nil, // domain is global
		Labels:       nil, // domain does not have labels,
		ParentId:     nil, // domain is the top-most item and have no parent,
		Raw:          collector.Raw(domain),
	}

	log.Info("Adding domain", slog.String("name", domain.Name))

	return r, nil
}

// handleProject returns a [ontology.ResourceGroup] out of an existing [projects.Project].
func (d *openstackCollector) handleProject(project *projects.Project) (ontology.IsResource, error) {
	r := &ontology.ResourceGroup{
		Id:          project.ID,
		Name:        project.Name,
		Description: project.Description,

		CreationTime: nil, // project does not have a creation date
		GeoLocation: &ontology.GeoLocation{
			Region: d.region,
		},
		Labels:   labels(util.Ref(project.Tags)),
		ParentId: util.Ref(project.ParentID),
		Raw:      collector.Raw(project),
	}

	log.Info("Adding project", slog.String("name", project.Name))

	return r, nil
}
