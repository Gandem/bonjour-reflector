package main

import (
	"net"
	"reflect"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	options gopacket.SerializeOptions
)

func createMockIPv4mDNSQueryPacket() []byte {
	var options gopacket.SerializeOptions
	srcMAC := net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	dstMAC := net.HardwareAddr{0xBD, 0xBD, 0xBD, 0xBD, 0xBD, 0xBD}
	srcIP := net.IP{127, 0, 0, 1}
	dstIP := net.IP{224, 0, 0, 251}
	srcPort := layers.UDPPort(5353)
	dstPort := layers.UDPPort(5353)
	dnsQuestions := []layers.DNSQuestion{layers.DNSQuestion{
		Name:  []byte("example.com"),
		Type:  layers.DNSTypeA,
		Class: layers.DNSClassIN,
	}}


	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(
		buffer,
		options,
		&layers.Ethernet{
			SrcMAC:       srcMAC,
			DstMAC:       dstMAC,
			EthernetType: layers.EthernetTypeDot1Q,
		},
		&layers.Dot1Q{
			Priority:       uint8(3),
			VLANIdentifier: uint16(30),
			Type:           layers.EthernetTypeIPv4,
		},
		&layers.IPv4{
			SrcIP:    srcIP,
			DstIP:    dstIP,
			Version:  4,
			Protocol: layers.IPProtocolUDP,
			Length:   146,
			IHL:      5,
			TOS:      0,
		},
		&layers.UDP{
			SrcPort: srcPort,
			DstPort: dstPort,
		},
		&layers.DNS{
			Questions: dnsQuestions,
		},
	)
	return buffer.Bytes()
}

func createMockIPv6mDNSPacket() []byte {
	var options gopacket.SerializeOptions
	srcMAC := net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	dstMAC := net.HardwareAddr{0xBD, 0xBD, 0xBD, 0xBD, 0xBD, 0xBD}
	srcIP := net.ParseIP("::1")
	dstIP := net.ParseIP("ff02::fb")
	srcPort := layers.UDPPort(5353)
	dstPort := layers.UDPPort(5353)
	dnsQuestions := []layers.DNSQuestion{layers.DNSQuestion{
		Name:  []byte("example.com"),
		Type:  layers.DNSTypeA,
		Class: layers.DNSClassIN,
	}}

	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(
		buffer,
		options,
		&layers.Ethernet{
			SrcMAC:       srcMAC,
			DstMAC:       dstMAC,
			EthernetType: layers.EthernetTypeDot1Q,
		},
		&layers.Dot1Q{
			Priority:       uint8(3),
			VLANIdentifier: uint16(30),
			Type:           layers.EthernetTypeIPv6,
		},
		&layers.IPv6{
			SrcIP:      srcIP,
			DstIP:      dstIP,
			Version:    6,
			Length:     48,
			NextHeader: layers.IPProtocolUDP,
		},
		&layers.UDP{
			SrcPort: srcPort,
			DstPort: dstPort,
		},
		&layers.DNS{
			Questions: dnsQuestions,
		},
	)
	return buffer.Bytes()
}

func createMockIPv4mDNSAnswerPacket() []byte {
	var options gopacket.SerializeOptions
	srcMAC := net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	dstMAC := net.HardwareAddr{0xBD, 0xBD, 0xBD, 0xBD, 0xBD, 0xBD}
	srcIP := net.IP{127, 0, 0, 1}
	dstIP := net.IP{224, 0, 0, 251}
	srcPort := layers.UDPPort(5353)
	dstPort := layers.UDPPort(5353)
	dnsAnswers := []layers.DNSResourceRecord{layers.DNSResourceRecord{
		Name:  []byte("example.com"),
		Type:  layers.DNSTypeA,
		Class: layers.DNSClassIN,
		TTL: 1024,
		IP: net.IP([]byte{1, 2, 3, 4}),
	}}


	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(
		buffer,
		options,
		&layers.Ethernet{
			SrcMAC:       srcMAC,
			DstMAC:       dstMAC,
			EthernetType: layers.EthernetTypeDot1Q,
		},
		&layers.Dot1Q{
			Priority:       uint8(3),
			VLANIdentifier: uint16(30),
			Type:           layers.EthernetTypeIPv4,
		},
		&layers.IPv4{
			SrcIP:    srcIP,
			DstIP:    dstIP,
			Version:  4,
			Protocol: layers.IPProtocolUDP,
			Length:   146,
			IHL:      5,
			TOS:      0,
		},
		&layers.UDP{
			SrcPort: srcPort,
			DstPort: dstPort,
		},
		&layers.DNS{
			Answers: dnsAnswers,
		},
	)
	return buffer.Bytes()
}

func TestParseEthernetLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockIPv4mDNSQueryPacket(), decoder, options)

	expectedResult := &net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	computedResult := parseEthernetLayer(packet)
	if !reflect.DeepEqual(expectedResult, computedResult) {
		t.Error("Error in parseEthernetLayer()")
	}
}

func TestParseVLANTag(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockIPv4mDNSQueryPacket(), decoder, options)

	result := &layers.Dot1Q{
		Priority:       uint8(3),
		DropEligible:   true,
		VLANIdentifier: uint16(30),
		Type:           layers.EthernetTypeIPv6,
	}

	expectedResult := &result.VLANIdentifier
	computedResult := parseVLANTag(packet)

	if !reflect.DeepEqual(expectedResult, computedResult) {
		t.Error("Error in parseEthernetLayer()")
	}
}

func TestParseIPLayer(t *testing.T) {
	options := gopacket.DecodeOptions{Lazy: true}
	decoder := gopacket.DecodersByLayerName["Ethernet"]

	packetIPv4 := gopacket.NewPacket(createMockIPv4mDNSQueryPacket(), decoder, options)

	expectedResultIPv4 := net.IP{224, 0, 0, 251}
	computedResultIPv4 := parseIPLayer(packetIPv4)
	if !reflect.DeepEqual(expectedResultIPv4, computedResultIPv4) {
		t.Error("Error in parseIPLayer() for IPv4 addresses")
	}

	packetIPv6 := gopacket.NewPacket(createMockIPv6mDNSPacket(), decoder, options)

	expectedResultIPv6 := net.ParseIP("ff02::fb")
	computedResultIPv6 := parseIPLayer(packetIPv6)
	if !reflect.DeepEqual(expectedResultIPv6, computedResultIPv6) {
		t.Error("Error in parseIPLayer() for IPv6 addresses")
	}
}

func TestParseUDPLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	var options gopacket.DecodeOptions
	packet := gopacket.NewPacket(createMockIPv4mDNSQueryPacket(), decoder, options)

	expectedResult := layers.UDPPort(5353)
	computedResult, _ := parseUDPLayer(packet)
	if !reflect.DeepEqual(expectedResult, computedResult) {
		t.Error("Error in parseUDPLayer()")
	}
}

func TestParseDNSPayload(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	var options gopacket.DecodeOptions
	
	packetQuestion := gopacket.NewPacket(createMockIPv4mDNSQueryPacket(), decoder, options)

	_, packetQuestionPayload := parseUDPLayer(packetQuestion)

	expectedResultQuestion := true
	computedResultQuestion := parseDNSPayload(packetQuestionPayload)
	if !reflect.DeepEqual(expectedResultQuestion, computedResultQuestion) {
		t.Error("Error in parseUDPLayer()")
	}

	packetAnswer := gopacket.NewPacket(createMockIPv4mDNSAnswerPacket(), decoder, options)

	_, packetAnswerPayload := parseUDPLayer(packetAnswer)

	expectedResultAnswer := true
	computedResultAnswer := parseDNSPayload(packetAnswerPayload)
	if !reflect.DeepEqual(expectedResultAnswer, computedResultAnswer) {
		t.Error("Error in parseUDPLayer()")
	}
}
