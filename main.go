package main

import (
	"fmt"
	"log"
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

	decoder := gopacket.DecodersByLayerName["Ethernet"]

	source := gopacket.NewPacketSource(rawTraffic, decoder)

	bonjourPackets := filterBonjourPacketsLazily(source)

	for bonjourPacket := range bonjourPackets {
		fmt.Println(bonjourPacket.packet.String())
		if bonjourPacket.vlanTag != nil {
			if bonjourPacket.isDNSQuery {
				if tags, ok := poolsMap[*bonjourPacket.vlanTag]; ok {
					for _, tag := range tags {
						*bonjourPacket.vlanTag = tag
						buf := gopacket.NewSerializeBuffer()
						gopacket.SerializePacket(buf, gopacket.SerializeOptions{}, bonjourPacket.packet)
						rawTraffic.WritePacketData(buf.Bytes())
					}
				}
			} else {
				// redirect to map by mac
			}
		}
	}
}
