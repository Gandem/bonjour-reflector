package main

import (
	"reflect"
	"testing"
	"sort"
)

var devices = map[macAddress]bonjourDevice{
  "00:14:22:01:23:45": bonjourDevice{OriginPool: 45, SharedPools: []int{42, 1042, 46}},
  "00:14:22:01:23:46": bonjourDevice{OriginPool: 46, SharedPools: []int{176, 148}},
  "00:14:22:01:23:47": bonjourDevice{OriginPool: 47, SharedPools: []int{1042, 1717, 13}},
}

func TestMapByPool(t *testing.T) {
  computedResult := mapByPool(devices)
	// Sort slices to ensure that a different order does not make the test fail
	for _, slice := range computedResult {
		sort.Ints(slice)
	}

  expectedResult := map[int]([]int){
    42: []int{45},
    1042: []int{45, 47},
    46: []int{45},
    176: []int{46},
    148: []int{46},
		13: []int{47},
    1717: []int{47},
  }
  if !reflect.DeepEqual(computedResult, expectedResult) {
    t.Error("Error in mapByPool()")
  }
}
