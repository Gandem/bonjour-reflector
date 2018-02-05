package main

import (
	"log"
	"net"
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

	log.Println(netInterface)
}
