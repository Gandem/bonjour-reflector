package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
	cfg, err := readConfig("./config.toml")
	if err != nil {
		log.Fatalf("Could not read configuration: %v", err)
	}
	poolsMap := mapByPool(cfg.Devices)

	rawTraffic, err := pcap.OpenLive(cfg.NetInterface, 65536, true, time.Second)
	if err != nil {
		log.Fatalf("Could not find network interface: %v", cfg.NetInterface)
	}

	// Get the Mac Address to filter for Bonjour packet duplicates
	intf, err := net.InterfaceByName(cfg.NetInterface)
	if err != nil {
		log.Fatal(err)
	}
	brMACAddress := intf.HardwareAddr

	decoder := gopacket.DecodersByLayerName["Ethernet"]

	source := gopacket.NewPacketSource(rawTraffic, decoder)

	bonjourPackets := filterBonjourPacketsLazily(source, brMACAddress)

	for bonjourPacket := range bonjourPackets {
		fmt.Println(bonjourPacket.packet.String())
		if bonjourPacket.vlanTag != nil {
			if bonjourPacket.isDNSQuery {
				if tags, ok := poolsMap[*bonjourPacket.vlanTag]; ok {
					for _, tag := range tags {
						*bonjourPacket.vlanTag = tag
						*bonjourPacket.srcMAC = brMACAddress
						sendBonjourPacket(rawTraffic, bonjourPacket.packet)
					}
				}
			} else {
				if device, ok := cfg.Devices[macAddress(bonjourPacket.srcMAC.String())]; ok {
					for _, tag := range device.SharedPools {
						*bonjourPacket.vlanTag = tag
						*bonjourPacket.srcMAC = brMACAddress
						sendBonjourPacket(rawTraffic, bonjourPacket.packet)
					}
				}
			}
		}
	}
}

func sendBonjourPacket(handle *pcap.Handle, packet gopacket.Packet) {
	buf := gopacket.NewSerializeBuffer()
	gopacket.SerializePacket(buf, gopacket.SerializeOptions{}, packet)
	handle.WritePacketData(buf.Bytes())
}
