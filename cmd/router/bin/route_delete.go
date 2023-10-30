package bin

import (
	"fmt"
	"github.com/Inspur-Data/router/pkg/modules"
	"github.com/Inspur-Data/router/pkg/config"
	"github.com/Inspur-Data/router/pkg/logging"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	//"time"
)

func CmdDel(args *skel.CmdArgs) (err error) {
	// for calu consuming  time
	//startTime := time.Now()
	logging.SetLogFile("/var/log/router-del.log")
	// parse args
	podArgs := config.PodArgs{}
	if err = types.LoadArgs(args.Args, &podArgs); nil != err {
		return fmt.Errorf("failed to load CNI ENV args: %w", err)
	}

	// parse router config info
	routerConfig := models.RouterConfig{}
	conf, err := ParseConfig(args.StdinData, &routerConfig)
	if err != nil || conf == nil {
		return err
	}



	return nil
}

/*
// DelRoute in ns
func DelRoute(ruleTable, ipFamily int, scope netlink.Scope, iface string, dst *net.IPNet, gw net.IP) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		logging.Errorf(err.Error())
		return err
	}
	// todo handle  table
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Scope:     scope,
		Dst:       dst,
		Gw:        gw,
		//Table:     ruleTable,
	}
	if err = netlink.RouteDel(route); err != nil && !os.IsExist(err) {
		logging.Errorf("failed to RouteDel,route:%v ,err:%v",route.String(),err.Error())
		return fmt.Errorf("failed to del route table(%v): %v", route.String(), err)
	}
	return nil
}
*/
