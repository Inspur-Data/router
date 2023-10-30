package bin

import (
	"encoding/json"
	"github.com/Inspur-Data/router/pkg/logging"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"os"
	"runtime/debug"

	"github.com/Inspur-Data/router/pkg/config"
	"github.com/containernetworking/cni/pkg/skel"
	"net"
)

var (
	defaultOverlayVethName = "eth0"
)

type RouterConfig struct {
	types.NetConf
	TableName          string       `json:"detectGateway,omitempty"`
	DetectGateway      *bool        `json:"detectGateway,omitempty"`
	ServiceCIDR        []string     `json:"serviceCIDR,omitempty"`
	PodDefaultRouteNIC string       `json:"podDefaultRouteNic,omitempty"`
	Routes             []RouterInfo `json:"routes"`
}
type RouterInfo struct {
	Dst string `json:"dst"`
	Gw  string `json:"dw"`
}

func CmdAdd(args *skel.CmdArgs) (err error) {
	// logging.Debugf("begin to CmdAdd args: %+v", args)
	logging.SetLogFile("/var/log/router-add.log")
	defer func() {
		if err := recover(); err != nil {
			logging.Errorf("Panic: %v\n%s", err, debug.Stack())
		}
	}()

	podArgs := config.PodArgs{}
	if err = types.LoadArgs(args.Args, &podArgs); nil != err {
		return logging.Errorf("failed to load CNI ENV args: %w", err)
	}
	logging.Debugf("podArgs :%v", podArgs)

	// 具体的路由配置信息,multus
	if len(args.StdinData) == 0 {
		logging.Debugf("args.StdinData is empty")
	} else {
		logging.Debugf("args.StdinData :%v", string(args.StdinData))
	}

	routerNetConf := config.RouterNetConf{}
	if err = json.Unmarshal(args.StdinData, &routerNetConf); err != nil {
		return logging.Errorf("failed to Unmarshal  routerConfig: %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	defer netns.Close()
	if err != nil {
		return logging.Errorf("failed to GetNS %q: %v", args.Netns, err)
	}
	logging.Debugf("netns :%v", netns)
	routerConfig := routerNetConf.Routes

	//globalDynamicRouterGW := routerConfig.DynamicRouterGW
	err = netns.Do(func(_ ns.NetNS) error {
		var v4Gw net.IP
		//var v6Gw net.IP
		logging.Debugf("ifName :%v", args.IfName)
		// get  ip by interface name
		/*link, err := netlink.LinkByName(args.IfName)
		if err != nil {
			return logging.Errorf("Failed to get link:", err.Error())
		}*/

		// 获取 IP 地址列表
		/*addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			return logging.Errorf("Failed to get address list:", err.Error())
		}

		if len(addrs) == 0 {
			logging.Debugf("ifName's ip is empty :%v", args.IfName)
			return nil
		}

		podArgs.IP = addrs[0].IP
		logging.Debugf("podArgs IP :%v", podArgs.IP)

		routers, err := netlink.RouteGet(podArgs.IP)
		if err != nil {
			return fmt.Errorf("failed to RouteGet Pod IP(%s): %v", podArgs.IP, err)
		}

		if len(routers) == 0 {
			logging.Debugf("podName-pod:%v-%v routers is empty", podArgs.K8S_POD_NAMESPACE, podArgs.K8S_POD_NAME)
			return nil
		}

		if podArgs.IP.To4() != nil && v4Gw == nil {
			v4Gw = routers[0].Src
		}*/

		// set rule in ns
		for _, route := range routerConfig {
			// get dst and gw
			_, dst, err := net.ParseCIDR(route.V4Dst)
			if err != nil {
				return logging.Errorf("failed to translate dst :%v,err: %v", dst, err.Error())
			}
			logging.Debugf("dst :%v", dst)

			dynamicRouterGW := route.DynamicRouterGW
			v4Gw = net.ParseIP(route.V4Gw)

			if !dynamicRouterGW {
				err = AddRoute(100, netlink.FAMILY_V4, netlink.SCOPE_UNIVERSE, args.IfName, dst, v4Gw)
				if err != nil {
					return logging.Errorf("failed to AddRoute : %v", err.Error())
				}
				continue
			}

			// dynamic Router GW
			// get all routes of current interface
			currentInterfaceRoutes, err := GetRoutesByName(args.IfName, netlink.FAMILY_V4)
			if err != nil {
				return logging.Errorf("failed to GetRoutesByName: %v", err.Error())
			}

			if len(currentInterfaceRoutes) == 0 {
				logging.Debugf("currentInterfaceRoutes len is zero")
				continue
			}

			for _, router := range currentInterfaceRoutes {
				logging.Debugf("failed to router from ifName router: %v", router.String())
				err = AddRoute(100, netlink.FAMILY_V4, netlink.SCOPE_UNIVERSE, args.IfName, dst, router.Gw)
				if err != nil {
					return logging.Errorf("failed to AddRoute : %v", err.Error())
				}
			}
		}
		logging.Debugf("netns.Do")
		return nil
	})
	if err = version.ParsePrevResult(&routerNetConf.NetConf); err != nil {
		return logging.Errorf("failed to parse prevResult: %v", err)
	}
	return types.PrintResult(routerNetConf.PrevResult, routerNetConf.CNIVersion)
}

// AddRoute add static route to specify rule table
func AddRoute(ruleTable, ipFamily int, scope netlink.Scope, iface string, dst *net.IPNet, gw net.IP) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		logging.Errorf(err.Error())
		return err
	}
	logging.Debugf(" link :%v", link.Type())
	// todo handle  table
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Scope:     scope,
		Dst:       dst,
		Gw:        gw,
		//Table:     ruleTable,
	}
	logging.Debugf(" route :%v", route.String())
	if err = netlink.RouteAdd(route); err != nil && !os.IsExist(err) {
		return logging.Errorf("failed to RouteAdd,route:%v ,err:%v", route.String(), err.Error())
	}
	return nil
}
