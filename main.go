package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	cfg, err := readConfig("./config.toml")
	if err != nil {
		log.Fatalf("Could not read configuration : %v", err)
	}

	src, err := pcap.OpenLive(cfg.NetInterface, 65536, true, time.Second)

	var dec gopacket.Decoder
	var ok bool

	if dec, ok = gopacket.DecodersByLayerName["Ethernet"]; !ok {
		log.Fatalln("No decoder named Ethernet")
	}
	source := gopacket.NewPacketSource(src, dec)

	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	var udp layers.UDP
	var tag layers.Dot1Q
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &tag, &eth, &ip4, &ip6, &udp)
	decoded := []gopacket.LayerType{}

	for packet := range source.Packets() {
		parser.DecodeLayers(packet.Data(), &decoded)
		// Detect Bonjour packets
		if ip4.DstIP.String() == "224.0.0.251" {
			if udp.SrcPort == 5353 && udp.DstPort == 5353 {
				fmt.Printf("New Bonjour packet detected from %v\n", ip4.SrcIP)
			}
		}
	}
}
