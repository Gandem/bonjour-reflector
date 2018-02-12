package main

import (
	"net"
	"reflect"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	srcMACTest = net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	dstMACTest = net.HardwareAddr{0xBD, 0xBD, 0xBD, 0xBD, 0xBD, 0xBD}
	vlanIdentifierTest = uint16(30)
	srcIPv4Test = net.IP{127, 0, 0, 1}
	dstIPv4Test = net.IP{224, 0, 0, 251}
	srcIPv6Test = net.ParseIP("::1")
	dstIPv6Test = net.ParseIP("ff02::fb")
	srcUDPPortTest = layers.UDPPort(5353)
	dstUDPPortTest = layers.UDPPort(5353)
)


func createMockmDNSPacket(isIPv4 bool, isDNSQuery bool) []byte {
	var ethernetLayer, dot1QLayer, ipLayer, udpLayer, dnsLayer gopacket.SerializableLayer

	ethernetLayer = &layers.Ethernet{
		SrcMAC:       srcMACTest,
		DstMAC:       dstMACTest,
		EthernetType: layers.EthernetTypeDot1Q,
	}

	if isIPv4 {
		dot1QLayer = &layers.Dot1Q{
			VLANIdentifier: vlanIdentifierTest,
			Type:           layers.EthernetTypeIPv4,
		}

		ipLayer = &layers.IPv4{
			SrcIP:    srcIPv4Test,
			DstIP:    dstIPv4Test,
			Version:  4,
			Protocol: layers.IPProtocolUDP,
			Length:   146,
			IHL:      5,
			TOS:      0,
		}
	} else {
		dot1QLayer = &layers.Dot1Q{
			VLANIdentifier: vlanIdentifierTest,
			Type:           layers.EthernetTypeIPv6,
		}

		ipLayer = &layers.IPv6{
			SrcIP:      srcIPv6Test,
			DstIP:      dstIPv6Test,
			Version:    6,
			Length:     48,
			NextHeader: layers.IPProtocolUDP,
		}
	}

	udpLayer = &layers.UDP{
		SrcPort: srcUDPPortTest,
		DstPort: dstUDPPortTest,
	}

	if isDNSQuery {
		dnsLayer = &layers.DNS{
			Questions: []layers.DNSQuestion{layers.DNSQuestion{
				Name:  []byte("example.com"),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
			}},
			QDCount: 1,
		}
	} else {
		dnsLayer = &layers.DNS{
			Answers: []layers.DNSResourceRecord{layers.DNSResourceRecord{
				Name:  []byte("example.com"),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
				TTL: 1024,
				IP: net.IP([]byte{1, 2, 3, 4}),
			}},
			ANCount: 1,
			QR: true,
		}
	}

	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{},
		ethernetLayer,
		dot1QLayer,
		ipLayer,
		udpLayer,
		dnsLayer,
	)
	return buffer.Bytes()
}

func TestParseEthernetLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockmDNSPacket(true, true), decoder, options)

	expectedResult := &srcMACTest
	computedResult := parseEthernetLayer(packet)
	if !reflect.DeepEqual(expectedResult, computedResult) {
		t.Error("Error in parseEthernetLayer()")
	}
}

func TestParseVLANTag(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockmDNSPacket(true, true), decoder, options)

	expectedLayer := &layers.Dot1Q{
		VLANIdentifier: vlanIdentifierTest,
		Type:           layers.EthernetTypeIPv4,
	}
	expectedResult := &expectedLayer.VLANIdentifier
	computedResult := parseVLANTag(packet)
	if !reflect.DeepEqual(expectedResult, computedResult) {
		t.Error("Error in parseEthernetLayer()")
	}
}

func TestParseIPLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	ipv4Packet := gopacket.NewPacket(createMockmDNSPacket(true, true), decoder, options)

	ipv4ExpectedResult := dstIPv4Test
	ipv4ComputedResult := parseIPLayer(ipv4Packet)
	if !reflect.DeepEqual(ipv4ExpectedResult, ipv4ComputedResult) {
		t.Error("Error in parseIPLayer() for IPv4 addresses")
	}

	ipv6Packet := gopacket.NewPacket(createMockmDNSPacket(false, true), decoder, options)

	ipv6ExpectedResult := dstIPv6Test
	ipv6ComputedResult := parseIPLayer(ipv6Packet)
	if !reflect.DeepEqual(ipv6ExpectedResult, ipv6ComputedResult) {
		t.Error("Error in parseIPLayer() for IPv6 addresses")
	}
}

func TestParseUDPLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockmDNSPacket(true, true), decoder, options)

	expectedResult := dstUDPPortTest
	computedResult, _ := parseUDPLayer(packet)
	if !reflect.DeepEqual(expectedResult, computedResult) {
		t.Error("Error in parseUDPLayer()")
	}
}

func TestParseDNSPayload(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	questionPacket := gopacket.NewPacket(createMockmDNSPacket(true, true), decoder, options)

	_, questionPacketPayload := parseUDPLayer(questionPacket)

	questionExpectedResult := true
	questionComputedResult := parseDNSPayload(questionPacketPayload)
	if !reflect.DeepEqual(questionExpectedResult, questionComputedResult) {
		t.Error("Error in parseDNSPayload() for DNS queries")
	}

	answerPacket := gopacket.NewPacket(createMockmDNSPacket(true, false), decoder, options)

	_, answerPacketPayload := parseUDPLayer(answerPacket)

	answerExpectedResult := false
	answerComputedResult := parseDNSPayload(answerPacketPayload)
	if !reflect.DeepEqual(answerExpectedResult, answerComputedResult) {
		t.Error("Error in parseDNSPayload() for DNS answers")
	}
}
