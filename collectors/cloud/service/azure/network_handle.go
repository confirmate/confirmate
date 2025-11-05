package azure

import (
	"clouditor.io/clouditor/v2/api/discovery"
	"clouditor.io/clouditor/v2/api/ontology"
	"clouditor.io/clouditor/v2/internal/util"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func (d *azureDiscovery) handleLoadBalancer(lb *armnetwork.LoadBalancer) ontology.IsResource {
	return &ontology.LoadBalancer{
		Id:           resourceID(lb.ID),
		Name:         util.Deref(lb.Name),
		CreationTime: nil, // No creation time available
		GeoLocation: &ontology.GeoLocation{
			Region: util.Deref(lb.Location),
		},
		Labels:   labels(lb.Tags),
		ParentId: resourceGroupID(lb.ID),
		Raw:      discovery.Raw(lb),
		Ips:      publicIPAddressFromLoadBalancer(lb),
		Ports:    loadBalancerPorts(lb), // TODO(oxisto): ports should be uint16, not 32
	}
}

// handleApplicationGateway returns the application gateway with its properties
// NOTE: handleApplicationGateway uses the LoadBalancer for now until there is a own resource
func (d *azureDiscovery) handleApplicationGateway(ag *armnetwork.ApplicationGateway) ontology.IsResource {
	firewallStatus := false

	if ag.Properties != nil && ag.Properties.WebApplicationFirewallConfiguration != nil {
		firewallStatus = util.Deref(ag.Properties.WebApplicationFirewallConfiguration.Enabled)
	}

	return &ontology.LoadBalancer{
		Id:           resourceID(ag.ID),
		Name:         util.Deref(ag.Name),
		CreationTime: nil, // No creation time available
		GeoLocation: &ontology.GeoLocation{
			Region: util.Deref(ag.Location),
		},
		Labels:   labels(ag.Tags),
		ParentId: resourceGroupID(ag.ID),
		Raw:      discovery.Raw(ag),
		AccessRestriction: &ontology.AccessRestriction{
			Type: &ontology.AccessRestriction_WebApplicationFirewall{
				WebApplicationFirewall: &ontology.WebApplicationFirewall{
					Enabled: firewallStatus,
				},
			},
		},
	}
}

func (d *azureDiscovery) handleNetworkInterfaces(ni *armnetwork.Interface) ontology.IsResource {
	return &ontology.NetworkInterface{
		Id:           resourceID(ni.ID),
		Name:         util.Deref(ni.Name),
		CreationTime: nil, // No creation time available
		GeoLocation: &ontology.GeoLocation{
			Region: util.Deref(ni.Location),
		},
		Labels:   labels(ni.Tags),
		ParentId: resourceGroupID(ni.ID),
		Raw:      discovery.Raw(ni),
		AccessRestriction: &ontology.AccessRestriction{
			Type: &ontology.AccessRestriction_L3Firewall{
				L3Firewall: &ontology.L3Firewall{
					Enabled: d.nsgFirewallEnabled(ni),
				},
			},
		},
	}
}
