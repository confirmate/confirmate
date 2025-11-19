package azure

import (
	"log/slog"

	"confirmate.io/core/api/ontology"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// discoverNetworkInterfaces discovers network interfaces
func (d *azureDiscovery) discoverNetworkInterfaces() ([]ontology.IsResource, error) {
	var list []ontology.IsResource

	// initialize network interfaces client
	if err := d.initNetworkInterfacesClient(); err != nil {
		return nil, err
	}

	// List all network interfaces
	err := listPager(d,
		d.clients.networkInterfacesClient.NewListAllPager,
		d.clients.networkInterfacesClient.NewListPager,
		func(res armnetwork.InterfacesClientListAllResponse) []*armnetwork.Interface {
			return res.Value
		},
		func(res armnetwork.InterfacesClientListResponse) []*armnetwork.Interface {
			return res.Value
		},
		func(ni *armnetwork.Interface) error {
			s := d.handleNetworkInterfaces(ni)

			slog.Info("Adding network interface", slog.String("network interface", s.GetName()))

			list = append(list, s)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}

// discoverApplicationGateway discovers application gateways
func (d *azureDiscovery) discoverApplicationGateway() ([]ontology.IsResource, error) {
	var list []ontology.IsResource

	// initialize application gateway client
	if err := d.initApplicationGatewayClient(); err != nil {
		return nil, err
	}

	// List all application gateways
	err := listPager(d,
		d.clients.applicationGatewayClient.NewListAllPager,
		d.clients.applicationGatewayClient.NewListPager,
		func(res armnetwork.ApplicationGatewaysClientListAllResponse) []*armnetwork.ApplicationGateway {
			return res.Value
		},
		func(res armnetwork.ApplicationGatewaysClientListResponse) []*armnetwork.ApplicationGateway {
			return res.Value
		},
		func(ags *armnetwork.ApplicationGateway) error {
			s := d.handleApplicationGateway(ags)

			slog.Info("Adding application gateway", slog.String("application gateway", s.GetName()))

			list = append(list, s)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}

// discoverLoadBalancer discovers load balancer
func (d *azureDiscovery) discoverLoadBalancer() ([]ontology.IsResource, error) {
	var list []ontology.IsResource

	// initialize load balancers client
	if err := d.initLoadBalancersClient(); err != nil {
		return nil, err
	}

	// List all load balancers
	err := listPager(d,
		d.clients.loadBalancerClient.NewListAllPager,
		d.clients.loadBalancerClient.NewListPager,
		func(res armnetwork.LoadBalancersClientListAllResponse) []*armnetwork.LoadBalancer {
			return res.Value
		},
		func(res armnetwork.LoadBalancersClientListResponse) []*armnetwork.LoadBalancer {
			return res.Value
		},
		func(lbs *armnetwork.LoadBalancer) error {
			s := d.handleLoadBalancer(lbs)

			slog.Info("Adding load balancer", slog.String("load balancer", s.GetName()))

			list = append(list, s)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}
