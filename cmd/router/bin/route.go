package bin

import (
	"encoding/json"
	"github.com/Inspur-Data/router/api/v1/models"
	"github.com/Inspur-Data/router/pkg/logging"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/vishvananda/netlink"
	"k8s.io/utils/pointer"
)

// GetRoutesByName return all routes is belonged to specify interface
// filter by family also
func GetRoutesByName(iface string, ipfamily int) (routes []netlink.Route, err error) {
	var link netlink.Link
	if iface != "" {
		link, err = netlink.LinkByName(iface)
		if err != nil {
			return nil, err
		}
	}

	return netlink.RouteList(link, ipfamily)
}

// ParseConfig parses the supplied configuration (and prevResult) from stdin.
func ParseConfig(stdin []byte, routerConfig *models.RouterConfig) (*RouterConfig, error) {
	var err error
	conf := RouterConfig{}

	if err = json.Unmarshal(stdin, &conf); err != nil {
		return nil, logging.Errorf("failed to parse config: %v", err)
	}

	if err = version.ParsePrevResult(&conf.NetConf); err != nil {
		return nil, logging.Errorf("failed to parse prevResult: %v", err)
	}
	/*
		if err = routerConfig.Validate(strfmt.Default); err != nil {
			return nil, err
		}
	*/
	if conf.PodDefaultRouteNIC == "" {
		conf.PodDefaultRouteNIC = defaultOverlayVethName
	}

	if conf.DetectGateway == nil {
		conf.DetectGateway = pointer.Bool(routerConfig.DetectGateway)
	}

	if conf.PodDefaultRouteNIC == "" && routerConfig.PodDefaultRouteNIC != "" {
		conf.PodDefaultRouteNIC = routerConfig.PodDefaultRouteNIC
	}

	if len(conf.Routes) == 0 {
		// if not have routes,we don't show any error
		return nil, nil
	}
	return &conf, nil
}
