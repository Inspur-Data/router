package config

import (
	"encoding/json"
	"github.com/Inspur-Data/router/pkg/constant"
	logging "github.com/Inspur-Data/router/pkg/logging"
)

// ParseConfig the args in []byte to CNIConf
var CniConfig = &CNIConf{}

func ParseConfig(args []byte) (*CNIConf, error) {

	err := json.Unmarshal(args, CniConfig)
	if err != nil {
		logging.Errorf("json unmarshal failed: %v", err)
		return nil, err
	}
	if CniConfig.IPAM.LogFile != "" {
		logging.SetLogFile(CniConfig.IPAM.LogFile)
	}
	/*else {
		logging.SetLogFile("/var/log/router.log")
	}*/

	if CniConfig.IPAM.LogLevel != "" {
		logging.SetLogLevel(CniConfig.IPAM.LogLevel)
	} /*else {
		logging.SetLogLevel("debug")
	}*/

	if CniConfig.IPAM == nil {
		return nil, logging.Errorf("IPAM config is nil")
	}
	if CniConfig.IPAM.UnixSocketPath == "" {
		CniConfig.IPAM.UnixSocketPath = constant.DefaultIPAMUnixSocketPath
	}
	for _, version := range SupportCniVersion {
		if CniConfig.CNIVersion == version {
			return CniConfig, nil
		}
	}
	return nil, logging.Errorf("unsupported cni version: %v,the supported cni version %v", CniConfig.CNIVersion, SupportCniVersion)
}
