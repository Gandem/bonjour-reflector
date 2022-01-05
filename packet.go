package main

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type bonjourPacket struct {
	packet     gopacket.Packet
	srcMAC     *net.HardwareAddr
	dstMAC     *net.HardwareAddr
	isIPv6     bool
	srcIP     *net.IP
	dstIP     *net.IP
	vlanTag    *uint16
	isDNSQuery bool
}

func parsePacketsLazily(source *gopacket.PacketSource) chan bonjourPacket {
	// Process packets, and forward Bonjour traffic to the returned channel

	// Set decoding to Lazy
	source.DecodeOptions = gopacket.DecodeOptions{Lazy: true}

	packetChan := make(chan bonjourPacket, 100)

	go func() {
		for packet := range source.Packets() {
			tag := parseVLANTag(packet)

			// Get source and destination mac addresses
			srcMAC, dstMAC := parseEthernetLayer(packet)
			
			// Check IP protocol version and get srcIP
			isIPv6, srcIP := parseIPLayer(packet)
			
			// Get UDP payload
			payload := parseUDPLayer(packet)

			isDNSQuery := parseDNSPayload(payload)

			// Pass on the packet for its next adventure
			packetChan <- bonjourPacket{
				packet:     packet,
				vlanTag:    tag,
				srcMAC:     srcMAC,
				dstMAC:     dstMAC,
				isIPv6:     isIPv6,
				srcIP:     srcIP,
				isDNSQuery: isDNSQuery,
			}
		}
	}()

	return packetChan
}

func parseEthernetLayer(packet gopacket.Packet) (srcMAC, dstMAC *net.HardwareAddr) {
	if parsedEth := packet.Layer(layers.LayerTypeEthernet); parsedEth != nil {
		srcMAC = &parsedEth.(*layers.Ethernet).SrcMAC
		dstMAC = &parsedEth.(*layers.Ethernet).DstMAC
	}
	return
}

func parseVLANTag(packet gopacket.Packet) (tag *uint16) {
	if parsedTag := packet.Layer(layers.LayerTypeDot1Q); parsedTag != nil {
		tag = &parsedTag.(*layers.Dot1Q).VLANIdentifier
	}
	return
}

func parseIPLayer(packet gopacket.Packet) (isIPv6 bool, srcIP *net.IP) {
	if parsedIP := packet.Layer(layers.LayerTypeIPv4); parsedIP != nil {
		isIPv6 = false
		srcIP = &parsedIP.(*layers.IPv4).SrcIP
	}
	if parsedIP := packet.Layer(layers.LayerTypeIPv6); parsedIP != nil {
		isIPv6 = true
		srcIP = &parsedIP.(*layers.IPv6).SrcIP
	}
	
	return
}

func parseUDPLayer(packet gopacket.Packet) (payload []byte) {
	if parsedUDP := packet.Layer(layers.LayerTypeUDP); parsedUDP != nil {
		payload = parsedUDP.(*layers.UDP).Payload
	}
	return
}

func parseDNSPayload(payload []byte) (isDNSQuery bool) {
	packet := gopacket.NewPacket(payload, layers.LayerTypeDNS, gopacket.Default)
	if parsedDNS := packet.Layer(layers.LayerTypeDNS); parsedDNS != nil {
		isDNSQuery = !parsedDNS.(*layers.DNS).QR
	}
	return
}

type packetWriter interface {
	WritePacketData([]byte) error
}

func sendBonjourPacket(
	handle packetWriter,
	bonjourPacket *bonjourPacket,
	tag uint16,
	brMACAddress net.HardwareAddr,
	srcIPAddress net.IP,
	spoofsrcIP bool,
	dstMACAddress net.HardwareAddr,
	spoofdstMAC bool) {
	*bonjourPacket.vlanTag = tag
	*bonjourPacket.srcMAC = brMACAddress
	

	
	// Network devices may set dstMAC to the local MAC address
	// Rewrite dstMAC to ensure that it is set to the appropriate multicast MAC address
	if bonjourPacket.isIPv6 {
		*bonjourPacket.dstMAC = net.HardwareAddr{0x33, 0x33, 0x00, 0x00, 0x00, 0xFB}
	} else if spoofdstMAC && bonjourPacket.isIPv6 == false {
		*bonjourPacket.dstMAC = dstMACAddress
	} else {
		*bonjourPacket.dstMAC = net.HardwareAddr{0x01, 0x00, 0x5E, 0x00, 0x00, 0xFB}
	}
	 
	
	buf := gopacket.NewSerializeBuffer()
	serializeOptions := gopacket.SerializeOptions{}
	
	// We change the Source IP address of the mDNS query since Chromecasts ignore
	// packets coming from outside their subnet.
	if spoofsrcIP && bonjourPacket.isIPv6 == false {
		serializeOptions = gopacket.SerializeOptions{ComputeChecksums: true}
		*bonjourPacket.srcIP = srcIPAddress
		// We recalculate the checksum since the IP was modified
		if parsedIP := bonjourPacket.packet.Layer(layers.LayerTypeIPv4); parsedIP != nil {
			if parsedUDP := bonjourPacket.packet.Layer(layers.LayerTypeUDP); parsedUDP != nil {
				parsedUDP.(*layers.UDP).SetNetworkLayerForChecksum(parsedIP.(*layers.IPv4))
			}
		}
	}
		
	gopacket.SerializePacket(buf, serializeOptions, bonjourPacket.packet)
	handle.WritePacketData(buf.Bytes())
}
