package main

import (
	"reflect"
	"sort"
	"testing"
)

var devices = map[macAddress]bonjourDevice{
	"00:14:22:01:23:45": {OriginPool: 45, SharedPools: []uint16{42, 1042, 46}},
	"00:14:22:01:23:46": {OriginPool: 46, SharedPools: []uint16{176, 148}},
	"00:14:22:01:23:47": {OriginPool: 47, SharedPools: []uint16{1042, 1717, 13}},
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
