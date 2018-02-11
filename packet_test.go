package main

import (
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func createMockmDNSPacket(isIPv4 bool, isDNSQuery bool) []byte {
	var options gopacket.SerializeOptions
	var ethernetLayer, dot1QLayer, ipLayer, udpLayer, dnsLayer gopacket.SerializableLayer

	ethernetLayer = &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA},
		DstMAC:       net.HardwareAddr{0xBD, 0xBD, 0xBD, 0xBD, 0xBD, 0xBD},
		EthernetType: layers.EthernetTypeDot1Q,
	}

	if isIPv4 {
		dot1QLayer = &layers.Dot1Q{
			Priority:       uint8(3),
			VLANIdentifier: uint16(30),
			Type:           layers.EthernetTypeIPv4,
		}

		ipLayer = &layers.IPv4{
			SrcIP:    net.IP{127, 0, 0, 1},
			DstIP:    net.IP{224, 0, 0, 251},
			Version:  4,
			Protocol: layers.IPProtocolUDP,
			Length:   146,
			IHL:      5,
			TOS:      0,
		}
	} else {
		dot1QLayer = &layers.Dot1Q{
			Priority:       uint8(3),
			VLANIdentifier: uint16(30),
			Type:           layers.EthernetTypeIPv6,
		}

		ipLayer = &layers.IPv6{
			SrcIP:      net.ParseIP("::1"),
			DstIP:      net.ParseIP("ff02::fb"),
			Version:    6,
			Length:     48,
			NextHeader: layers.IPProtocolUDP,
		}
	}

	udpLayer = &layers.UDP{
		SrcPort: layers.UDPPort(5353),
		DstPort: layers.UDPPort(5353),
	}

	if isDNSQuery {
		dnsLayer = &layers.DNS{
			Questions: []layers.DNSQuestion{layers.DNSQuestion{
				Name:  []byte("example.com"),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
			}},
		}
	} else {
		dnsLayer = &layers.DNS{
			Answers: []layers.DNSResourceRecord{layers.DNSResourceRecord{
				Name:  []byte("example.com"),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
				TTL:   1024,
				IP:    net.IP([]byte{1, 2, 3, 4}),
			}},
		}
	}

	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(
		buffer,
		options,
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

	expectedResult := &net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
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
		Priority:       uint8(3),
		DropEligible:   true,
		VLANIdentifier: uint16(30),
		Type:           layers.EthernetTypeIPv6,
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

	ipv4ExpectedResult := net.IP{224, 0, 0, 251}
	ipv4ComputedResult := parseIPLayer(ipv4Packet)
	if !reflect.DeepEqual(ipv4ExpectedResult, ipv4ComputedResult) {
		t.Error("Error in parseIPLayer() for IPv4 addresses")
	}

	ipv6Packet := gopacket.NewPacket(createMockmDNSPacket(false, true), decoder, options)

	ipv6ExpectedResult := net.ParseIP("ff02::fb")
	ipv6ComputedResult := parseIPLayer(ipv6Packet)
	if !reflect.DeepEqual(ipv6ExpectedResult, ipv6ComputedResult) {
		t.Error("Error in parseIPLayer() for IPv6 addresses")
	}
}

func TestParseUDPLayer(t *testing.T) {
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	options := gopacket.DecodeOptions{Lazy: true}

	packet := gopacket.NewPacket(createMockmDNSPacket(true, true), decoder, options)

	expectedResult := layers.UDPPort(5353)
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
		t.Error("Error in parseUDPLayer()")
	}

	answerPacket := gopacket.NewPacket(createMockmDNSPacket(true, false), decoder, options)

	_, answerPacketPayload := parseUDPLayer(answerPacket)

	answerExpectedResult := true
	answerComputedResult := parseDNSPayload(answerPacketPayload)
	if !reflect.DeepEqual(answerExpectedResult, answerComputedResult) {
		t.Error("Error in parseUDPLayer()")
	}
}

type dataSource struct {
	packetSent bool
	data       []byte
}

func (dataSource *dataSource) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	// Return one packet.
	// If a packet has already been returned in the past, return an EOF error
	// to end the reading of packets from this source.
	data = dataSource.data
	ci = gopacket.CaptureInfo{
		Timestamp:      time.Time{},
		CaptureLength:  len(data),
		Length:         ci.CaptureLength,
		InterfaceIndex: 0,
	}
	if !dataSource.packetSent {
		dataSource.packetSent = true
		return data, ci, nil
	}
	return nil, ci, io.EOF
}

func createMockPacketSource() (packetSource *gopacket.PacketSource, packet gopacket.Packet) {
	data := createMockmDNSPacket(true, true)
	dataSource := &dataSource{
		packetSent: false,
		data:       data,
	}
	decoder := gopacket.DecodersByLayerName["Ethernet"]
	packetSource = gopacket.NewPacketSource(dataSource, decoder)
	packet = gopacket.NewPacket(data, decoder, gopacket.DecodeOptions{Lazy: true})
	return
}

func areBonjourPacketsEqual(a, b bonjourPacket) (areEqual bool) {
	areEqual = *a.vlanTag == *b.vlanTag && a.srcMAC.String() == b.srcMAC.String() && a.isDNSQuery == b.isDNSQuery
	// While comparing Bonjour packets, we do not want to compare packets entirely.
	// In particular, packet.metadata may be slightly different, we do not need them to be the same.
	// So we only compare the layers part of the packets.
	areEqual = areEqual && reflect.DeepEqual(a.packet.Layers(), b.packet.Layers())
	return
}

func TestFilterBonjourPacketsLazily(t *testing.T) {
	mockPacketSource, packet := createMockPacketSource()

	brMACaddress := net.HardwareAddr{0xF2, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	packetChan := filterBonjourPacketsLazily(mockPacketSource, brMACaddress)

	vlanTag := uint16(30)
	srcMAC := net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA}
	expectedResult := bonjourPacket{
		packet:     packet,
		vlanTag:    &vlanTag,
		srcMAC:     &srcMAC,
		isDNSQuery: true,
	}

	computedResult := <-packetChan
	if !areBonjourPacketsEqual(expectedResult, computedResult) {
		t.Error("Error in filterBonjourPacketsLazily()")
	}
}
