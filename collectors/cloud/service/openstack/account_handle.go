package openstack

import (
	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
)

// handleDomain returns a [ontology.Account] out of an existing [domains.Domain].
func (d *openstackDiscovery) handleDomain(domain *domains.Domain) (ontology.IsResource, error) {
	r := &ontology.Account{
		Id:           domain.ID,
		Name:         domain.Name,
		Description:  domain.Description,
		CreationTime: nil, // domain does not have a creation date
		GeoLocation:  nil, // domain is global
		Labels:       nil, // domain does not have labels,
		ParentId:     nil, // domain is the top-most item and have no parent,
		Raw:          cloud.Raw(domain),
	}

	log.Infof("Adding domain '%s", domain.Name)

	return r, nil
}

// handleProject returns a [ontology.ResourceGroup] out of an existing [projects.Project].
func (d *openstackDiscovery) handleProject(project *projects.Project) (ontology.IsResource, error) {
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
		Raw:      cloud.Raw(project),
	}

	log.Infof("Adding project '%s", project.Name)

	return r, nil
}
