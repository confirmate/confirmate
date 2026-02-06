package azure

import (
	"log/slog"

	"confirmate.io/core/api/ontology"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// collectNetworkInterfaces collects network interfaces
func (d *azureCollector) collectNetworkInterfaces() ([]ontology.IsResource, error) {
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

			log.Info("Adding network interface", slog.String("network interface", s.GetName()))

			list = append(list, s)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}

// collectApplicationGateway collects application gateways
func (d *azureCollector) collectApplicationGateway() ([]ontology.IsResource, error) {
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

			log.Info("Adding application gateway", slog.String("application gateway", s.GetName()))

			list = append(list, s)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}

// collectLoadBalancer collects load balancer
func (d *azureCollector) collectLoadBalancer() ([]ontology.IsResource, error) {
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

			log.Info("Adding load balancer", slog.String("load balancer", s.GetName()))

			list = append(list, s)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return list, nil
}
