package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type macAddress string

type brconfig struct {
	NetInterface string                       `toml:"net_interface"`
	Devices      map[macAddress]bonjourDevice `toml:"devices"`
}

type bonjourDevice struct {
	OriginPool  uint16   `toml:"origin_pool"`
	SharedPools []uint16 `toml:"shared_pools"`
}

func readConfig(path string) (cfg brconfig, err error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return brconfig{}, err
	}
	_, err = toml.Decode(string(content), &cfg)
	return cfg, err
}

func mapByPool(devices map[macAddress]bonjourDevice) map[uint16]([]uint16) {
	seen := make(map[uint16]map[uint16]bool)
	poolsMap := make(map[uint16]([]uint16))
	for _, device := range devices {
		for _, pool := range device.SharedPools {
			if _, ok := seen[pool]; !ok {
				seen[pool] = make(map[uint16]bool)
			}
			if _, ok := seen[pool][device.OriginPool]; !ok {
				seen[pool][device.OriginPool] = true
				poolsMap[pool] = append(poolsMap[pool], device.OriginPool)
			}
		}
	}
	return poolsMap
}
