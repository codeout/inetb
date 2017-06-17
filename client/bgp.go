package client

import (
	"errors"
	"fmt"
	"log"
	"github.com/osrg/gobgp/packet/bgp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)


func (c *Client) ReadBGPUpdate(direction int, ch chan *bgp.BGPUpdate) error {
	neighbor, err := c.Neighbor()
	if err != nil {
		return err
	}
	iface, err := c.PeerInterface()
	if err != nil {
		return err
	}

	var filter string
	switch direction {
	case Export:
		log.Printf("Start capturing outgoing BGP updates from %s on \"%s\"", neighbor.Config.NeighborAddress, iface)
		filter = fmt.Sprintf("tcp and port 179 and dst %s", neighbor.Config.NeighborAddress)
	case Import:
		log.Printf("Start capturing incoming BGP updates to %s on \"%s\"", neighbor.Config.NeighborAddress, iface)
		filter = fmt.Sprintf("tcp and port 179 and src %s", neighbor.Config.NeighborAddress)
	default:
		return errors.New("Unknown direction")
	}

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
