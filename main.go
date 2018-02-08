package main

import (
	"errors"
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
		log.Fatalf("Could not read configuration: %v", err)
	}

	src, err := pcap.OpenLive(cfg.NetInterface, 65536, true, time.Second)
	if err != nil {
		log.Fatalf("Could not find network interface: %v", cfg.NetInterface)
	}

	var dec gopacket.Decoder
	var ok bool

	if dec, ok = gopacket.DecodersByLayerName["Ethernet"]; !ok {
		log.Fatalln("No decoder named Ethernet")
	}
	source := gopacket.NewPacketSource(src, dec)

	for packet := range source.Packets() {
		// // Extract DOT1Q tag
		// var tag *layers.Dot1Q
		// if parsedTag := packet.Layer(layers.LayerTypeDot1Q); parsedTag != nil {
		// 	tag, _ = parsedTag.(*layers.Dot1Q)
		// }

		var ip4 layers.IPv4
		if parsedIP := packet.Layer(layers.LayerTypeIPv4); parsedIP != nil {
			ip4 = *parsedIP.(*layers.IPv4)
		}

		var ip6 layers.IPv6
		if parsedIP := packet.Layer(layers.LayerTypeIPv6); parsedIP != nil {
			ip6 = *parsedIP.(*layers.IPv6)
		}

		var udp layers.UDP
		if parsedUDP := packet.Layer(layers.LayerTypeUDP); parsedUDP != nil {
			udp = *parsedUDP.(*layers.UDP)
		}

		// Detect Bonjour packets
		if ip4.DstIP.String() == "224.0.0.251" || ip6.DstIP.String() == "ff02::fb" {
			if udp.DstPort == 5353 {
				// Print time for logging / debugging purposes
				fmt.Printf("[%v] New Bonjour packet detected from %v\n",
					time.Now().Format("02/01/2006 15:04:05"), ip4.SrcIP.String()) // Custom time layouts must use the reference time: Mon Jan 2 15:04:05 MST 2006
				dns, err := parseDNSPacket(udp.Payload)
				if err != nil {
					log.Println(err)
				}
				fmt.Println(formatDNSPacket(dns))
			}
		}
	}
}

func parseDNSPacket(payload []byte) (layers.DNS, error) {
	dnsParsedPacket := gopacket.NewPacket(payload, layers.LayerTypeDNS, gopacket.Default)
	var dns layers.DNS
	if parsedDNS := dnsParsedPacket.Layer(layers.LayerTypeDNS); parsedDNS != nil {
		dns = *parsedDNS.(*layers.DNS)
		return dns, nil
	}
	return dns, errors.New("Could not parse dns packet")
}

func formatDNSPacket(dns layers.DNS) (res string) {
	for _, answer := range dns.Answers {
		res += answer.String() + "\n"
	}
	for _, question := range dns.Questions {
		res += fmt.Sprintf("Question: %s %s %s \n", question.Class, question.Name, question.Type)
	}
	return res
}
