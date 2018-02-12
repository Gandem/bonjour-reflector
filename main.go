package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

const configPath = "./config.toml"

func main() {
	// Read config file and generate mDNS forwarding maps
	cfg, err := readConfig(configPath)
	if err != nil {
		log.Fatalf("Could not read configuration: %v", err)
	}
	poolsMap := mapByPool(cfg.Devices)

	// Get a handle on the network interface
	rawTraffic, err := pcap.OpenLive(cfg.NetInterface, 65536, true, time.Second)
	if err != nil {
		log.Fatalf("Could not find network interface: %v", cfg.NetInterface)
	}

	// Get the local MAC address, to filter out Bonjour packet generated locally
	intf, err := net.InterfaceByName(cfg.NetInterface)
	if err != nil {
		log.Fatal(err)
	}
	brMACAddress := intf.HardwareAddr

	// Get a channel of Bonjour packets to process
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	source := gopacket.NewPacketSource(rawTraffic, decoder)
	bonjourPackets := filterBonjourPacketsLazily(source, brMACAddress)

	// Process Bonjours packets
	for bonjourPacket := range bonjourPackets {
		fmt.Println(bonjourPacket.packet.String())
		if bonjourPacket.vlanTag != nil {
			// Forward the mDNS query or response to appropriate VLANs
			if bonjourPacket.isDNSQuery {
				if tags, ok := poolsMap[*bonjourPacket.vlanTag]; ok {
					for _, tag := range tags {
						sendBonjourPacket(rawTraffic, &bonjourPacket, tag, brMACAddress)
					}
				}
			} else {
				if device, ok := cfg.Devices[macAddress(bonjourPacket.srcMAC.String())]; ok {
					for _, tag := range device.SharedPools {
						sendBonjourPacket(rawTraffic, &bonjourPacket, tag, brMACAddress)
					}
				}
			}
		}
	}
}

func sendBonjourPacket(handle *pcap.Handle, bonjourPacket *bonjourPacket, tag uint16, brMACAddress net.HardwareAddr) {
	*bonjourPacket.vlanTag = tag
	*bonjourPacket.srcMAC = brMACAddress
	buf := gopacket.NewSerializeBuffer()
	gopacket.SerializePacket(buf, gopacket.SerializeOptions{}, bonjourPacket.packet)
	handle.WritePacketData(buf.Bytes())
}
