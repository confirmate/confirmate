package openstack

import (
	"fmt"

	"confirmate.io/collectors/cloud/api/discovery"
	"confirmate.io/collectors/cloud/api/ontology"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
)

// discoverDomains discovers domains.
func (d *openstackDiscovery) discoverDomains() (list []ontology.IsResource, err error) {
	var opts domains.ListOptsBuilder = &domains.ListOpts{}
	list, err = genericList(d, d.identityClient, domains.List, d.handleDomain, domains.ExtractDomains, opts)

	// if we cannot retrieve the domain information by calling the API or from the environment variables, we will add the information manually if we already got the domain ID and/or domain name
	if err != nil {
		log.Debugf("Could not discover domains due to insufficient permissions, but we can proceed with less domain information: %v", err)

		if d.domain.domainID == "" {
			err := fmt.Errorf("domain ID is not available: %v", err)
			return nil, err
		}

		r := &ontology.Account{
			Id:   d.domain.domainID,
			Name: d.domain.domainName,
			Raw:  discovery.Raw("Domain information manually added."),
		}

		list = append(list, r)
	}

	return list, nil
}

// discoverProjects discovers projects/tenants. OpenStack project and tenant are interchangeable.
func (d *openstackDiscovery) discoverProjects() (list []ontology.IsResource, err error) {
	var opts projects.ListOptsBuilder = &projects.ListOpts{}
	list, err = genericList(d, d.identityClient, projects.List, d.handleProject, projects.ExtractProjects, opts)

	// if we cannot retrieve the project information by calling the API or from the environment variables, we will add the information manually if we already got the project ID and/or project name
	if err != nil {
		log.Debugf("Could not discover projects/tenants due to insufficient permissions, but we can proceed with less project/tenant information: %v", err)

		if d.project.projectID == "" {
			err := fmt.Errorf("domain ID is not available: %v", err)
			return nil, err
		}

		r := &ontology.ResourceGroup{
			Id:       d.project.projectID,
			Name:     d.project.projectName,
			ParentId: &d.domain.domainID,
			Raw:      discovery.Raw("Project/Tenant information manually added."),
		}

		list = append(list, r)
	}

	return list, nil
}
