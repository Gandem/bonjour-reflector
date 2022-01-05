package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
	// Read config file and generate mDNS forwarding maps
	configPath := flag.String("config", "", "Config file in TOML format")
	debug := flag.Bool("debug", false, "Enable pprof server on /debug/pprof/")
	flag.Parse()

	// Start debug server
	if *debug {
		go debugServer(6060)
	}

	cfg, err := readConfig(*configPath)
	if err != nil {
		log.Fatalf("Could not read configuration: %v", err)
	}
	poolsMap := mapByPool(cfg.Devices)

	// Get a handle on the network interface
	rawTraffic, err := pcap.OpenLive(cfg.NetInterface, 65536, true, time.Second)
	if err != nil {
		log.Fatalf("Could not find network interface: %v", cfg.NetInterface)
	}
	
	// Parse IP use to relay queries to chromecasts
	spoofAddr := net.ParseIP(cfg.SpoofAddr)
	if spoofAddr == nil {
		log.Fatalf("Could not parse cc_subnet_ip")
	}

	// Get the local MAC address, to filter out Bonjour packet generated locally
	intf, err := net.InterfaceByName(cfg.NetInterface)
	if err != nil {
		log.Fatal(err)
	}
	brMACAddress := intf.HardwareAddr

	// Filter tagged bonjour traffic
	filterTemplate := "not (ether src %s) and vlan and dst net (224.0.0.251 or ff02::fb) and udp dst port 5353"
	err = rawTraffic.SetBPFFilter(fmt.Sprintf(filterTemplate, brMACAddress))
	if err != nil {
		log.Fatalf("Could not apply filter on network interface: %v", err)
	}

	// Get a channel of Bonjour packets to process
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	source := gopacket.NewPacketSource(rawTraffic, decoder)
	bonjourPackets := parsePacketsLazily(source)

	// Map for the vlan to last MAC query
	lastquery := make(map[uint16]net.HardwareAddr)

	// Process Bonjours packets
	for bonjourPacket := range bonjourPackets {
		//fmt.Println(bonjourPacket.packet.String())

		// Forward the mDNS query or response to appropriate VLANs
		if bonjourPacket.isDNSQuery {
			// We store the MAC of the last client that sent a query so we can send the response directly to it
			lastquery[*bonjourPacket.vlanTag]=*bonjourPacket.srcMAC
			fmt.Printf("Storing MAC %v for vlan %v \n", *bonjourPacket.srcMAC, *bonjourPacket.vlanTag)
			tags, ok := poolsMap[*bonjourPacket.vlanTag]
			if !ok {
				continue
			}

			for _, tag := range tags {
				sendBonjourPacket(rawTraffic, &bonjourPacket, tag, brMACAddress, spoofAddr, true, *bonjourPacket.dstMAC, false)
			}
		} else {
			device, ok := cfg.Devices[macAddress(bonjourPacket.srcMAC.String())]
			if !ok {
				continue
			}
			for _, tag := range device.SharedPools {
				// if we have a MAC stored for this vlan we also send the response packet directly to it
				if clientMAC, ok := lastquery[tag]; ok {
					fmt.Printf("Sending direct packet to MAC %v \n", clientMAC)
					sendBonjourPacket(rawTraffic, &bonjourPacket, tag, brMACAddress, spoofAddr, false, clientMAC, true)
				}
				// we always forward the multicast answer
				sendBonjourPacket(rawTraffic, &bonjourPacket, tag, brMACAddress, spoofAddr, false, *bonjourPacket.dstMAC, false)
			}
		}
	}
}

func debugServer(port int) {
	err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil)
	if err != nil {
		log.Fatalf("The application was started with -debug flag but could not listen on port %v: \n %s", port, err)
	}
}
