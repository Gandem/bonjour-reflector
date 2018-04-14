package main

import (
	"os"
	"reflect"
	"sort"
	"testing"
)

var devices = map[macAddress]bonjourDevice{
	"00:14:22:01:23:45": bonjourDevice{OriginPool: 45, SharedPools: []uint16{42, 1042, 46}},
	"00:14:22:01:23:46": bonjourDevice{OriginPool: 46, SharedPools: []uint16{176, 148}},
	"00:14:22:01:23:47": bonjourDevice{OriginPool: 47, SharedPools: []uint16{1042, 1717, 13}},
}

func TestReadConfig(t *testing.T) {
	// Check that a valid config file is read adequately
	validTestConfigFile := "config_test.toml"
	computedCfg, err := readConfig(validTestConfigFile)
	expectedCfg := brconfig{
		NetInterface: "test0",
		Devices:      devices,
	}

	if err != nil {
		t.Errorf("Error in readConfig(): failed to read test config file %s", validTestConfigFile)
	} else if !reflect.DeepEqual(expectedCfg, computedCfg) {
		t.Error("Error in readConfig(): expected config does not match computed config")
	}

	// Check that a non-existant config file is handled adequately
	nonexistantConfigFile := "nonexistant_test.toml"
	computedCfg, err = readConfig(nonexistantConfigFile)
	if !reflect.DeepEqual(computedCfg, brconfig{}) {
		t.Error("Error in readConfig(): unexpected config returned for non-existant config file")
	}
	if !os.IsNotExist(err) {
		// if the error returned is not of type "file not found"
		t.Error("Error in readConfig(): wrong error returned for nonexistant config file")
	}
}

func TestMapByPool(t *testing.T) {
	computedResult := mapByPool(devices)
	// Sort slices to ensure that a different order does not make the test fail
	for _, slice := range computedResult {
		sort.Slice(slice, func(i, j int) bool { return slice[i] < slice[j] })
	}

	expectedResult := map[uint16]([]uint16){
		42:   []uint16{45},
		1042: []uint16{45, 47},
		46:   []uint16{45},
		176:  []uint16{46},
		148:  []uint16{46},
		13:   []uint16{47},
		1717: []uint16{47},
	}
	if !reflect.DeepEqual(computedResult, expectedResult) {
		t.Error("Error in mapByPool()")
	}
}
