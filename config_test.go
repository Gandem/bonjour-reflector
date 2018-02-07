package main

import (
	"reflect"
	"testing"
)

func TestAppendWithoutDuplicates(t *testing.T) {
  array1, array2 := []int{1, 2, 3}, []int{4, 5, 6, 3, 7}
  computedResult := appendWithoutDuplicates(array1, array2)
  expectedResult := []int{1, 2, 3, 4, 5, 6, 7}
  if !reflect.DeepEqual(computedResult, expectedResult) {
    t.Error("Error in appendWithoutDuplicates()")
  }
}

var devices = []bonjourDevice{
  bonjourDevice{MacAddress: "00:14:22:01:23:45", Pools: []int{42, 1042, 1337}},
  bonjourDevice{MacAddress: "00:14:22:01:23:46", Pools: []int{176, 148}},
  bonjourDevice{MacAddress: "00:14:22:01:23:47", Pools: []int{1042, 1717, 13}},
}

func TestMapByPool(t *testing.T) {
  computedResult := mapByPool(devices)
  expectedResult := map[int]([]int){
    42: []int{1042, 1337},
    1042: []int{42, 1337, 1717, 13},
    1337: []int{42, 1042},
    176: []int{148},
    148: []int{176},
    1717: []int{1042, 13},
    13: []int{1042, 1717},
  }
  if !reflect.DeepEqual(computedResult, expectedResult) {
    t.Error("Error in mapByPool()")
  }
}

func TestMapByAddress(t *testing.T) {
	computedResult := mapByAddress(devices)
	expectedResult := map[string]([]int){
    "00:14:22:01:23:45": []int{42, 1042, 1337},
    "00:14:22:01:23:46": []int{176, 148},
    "00:14:22:01:23:47": []int{1042, 1717, 13},
  }
	if !reflect.DeepEqual(computedResult, expectedResult) {
		t.Error("Error in mapByAddress()")
	}
}
