package main

import (
	"fmt"
	"log"
	"net"

	"github.com/songgao/ether"
	"github.com/songgao/packets/ethernet"
)

func main() {
	cfg, err := readConfig("./config.toml")
	if err != nil {
		log.Fatalf("Could not read configuration : %v", err)
	}

	netInterface, err := net.InterfaceByName(cfg.NetInterface)
	if err != nil {
		log.Fatalf("Could not list interfaces: %s", err)
	}

	netDevice, err := ether.NewDev(netInterface, nil)
	if err != nil {
		log.Fatalf("Could not tap to interface: %s", err)
	}

	var res ethernet.Frame
	netDevice.Read(&res)

	fmt.Println(res)
}
