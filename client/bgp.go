package client

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/osrg/gobgp/packet/bgp"
	"log"
)

func (c *Client) StartReader() error {
	neighbor, err := c.Neighbor()
	if err != nil {
		return err
	}
	iface, err := c.PeerInterface()
	if err != nil {
		return err
	}

	log.Printf("Start capturing outgoing BGP updates from %s on \"%s\"", neighbor.Config.NeighborAddress, iface)
	filter := fmt.Sprintf("tcp and port 179 and host %s", neighbor.Config.NeighborAddress)

	handle, err := pcap.OpenLive(iface, 1600, false, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	if err = handle.SetBPFFilter(filter); err != nil {
		log.Fatal(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		tcpLayer := packet.Layer(layers.LayerTypeTCP)

		if tcpLayer != nil {
			tcp, ok := tcpLayer.(*layers.TCP)
			if !ok {
				continue
			}

			msg, err := bgp.ParseBGPMessage(tcp.Payload)
			if err != nil {
				continue
			}

			if msg.Header.Type == bgp.BGP_MSG_UPDATE {
				ch <- msg.Body.(*bgp.BGPUpdate)
			}
		}
	}

	return nil
}
