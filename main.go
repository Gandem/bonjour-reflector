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
	var tcp layers.TCP
	var tag layers.Dot1Q
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &tag, &eth, &ip4, &ip6, &tcp)
	decoded := []gopacket.LayerType{}

	for packet := range source.Packets() {
		parser.DecodeLayers(packet.Data(), &decoded)
		for _, layerType := range decoded {
			switch layerType {
			case layers.LayerTypeDot1Q:
				fmt.Println(tag.VLANIdentifier)
			case layers.LayerTypeEthernet:
				fmt.Printf("Source MAC: %v, Dest MAC: %v \n", eth.DstMAC, eth.SrcMAC)
			case layers.LayerTypeIPv4:
				fmt.Printf("Source IP: %v, Dest IP: %v \n", ip4.SrcIP, ip4.DstIP)
			}
		}
	}
}
