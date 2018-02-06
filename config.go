package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type brconfig struct {
	NetInterface string          `toml:"net_interface"`
	Devices      []bonjourDevice `toml:"devices"`
}

type bonjourDevice struct {
	MacAddress string `toml:"mac_address"`
	Pools      []int  `toml:"pools"`
}

func readConfig(path string) (cfg brconfig, err error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return brconfig{}, err
	}
	_, err = toml.Decode(string(content), &cfg)
	return cfg, err
}

func appendWithoutDuplicates(array1 []int, array2 []int) (result []int) {
	seen := make(map[int]bool)
	for _, num := range array1 {
		seen[num] = true
		result = append(result, num)

	}
	for _, num := range array2 {
		if _, ok := seen[num]; !ok {
			seen[num] = true
			result = append(result, num)
		}
	}
	return result
}

func mapAllPools(devices []bonjourDevice) map[int]([]int) {
	poolsMap := make(map[int]([]int))
	for _, device := range devices {
		for i, pool := range device.Pools {
			sharedWithPools := make([]int, len(device.Pools)-1)
			copy(sharedWithPools, device.Pools[:i])
			copy(sharedWithPools[i:], device.Pools[i+1:])
			poolsMap[pool] = appendWithoutDuplicates(poolsMap[pool], sharedWithPools)
		}
	}
	return poolsMap
}
