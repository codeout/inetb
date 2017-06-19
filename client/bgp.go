package client

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
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

	streamFactory := &bgpStreamFactory{
		updates: c.Updates,
	}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		tcpLayer := packet.Layer(layers.LayerTypeTCP)

		if tcpLayer != nil {
			tcp, ok := tcpLayer.(*layers.TCP)
			if !ok {
				continue
			}

			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
		}
	}

	return nil
}

type bgpStreamFactory struct {
	updates chan *bgp.BGPUpdate
}

type bgpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
	updates        chan *bgp.BGPUpdate
}

func (factory *bgpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	bstream := &bgpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		updates:   factory.updates,
	}
	go bstream.run()

	return &bstream.r
}

var BGP_MESSAGE_MARKER = bytes.Repeat([]byte{255}, 16)

func (b *bgpStream) split(data []byte, atEOF bool) (int, []byte, error) {
	start := 0
	markerLen := len(BGP_MESSAGE_MARKER)

	// find 0xff
	for ; start < len(data); start++ {
		if data[start] == BGP_MESSAGE_MARKER[0] {
			break
		}
	}

	// find BGP Message Marker
	for ; start <= len(data)-markerLen; start++ {
		if bytes.Equal(data[start:markerLen+start], BGP_MESSAGE_MARKER) {
			break
		}
	}

	// Request more data
	if start+markerLen+2 > len(data) {
		return 0, nil, nil
	}
	msgLen := int(binary.BigEndian.Uint16(data[start+markerLen : start+markerLen+2]))

	// Request more data
	if start+msgLen > len(data) {
		return 0, nil, nil
	}

	return start + msgLen, data[start : start+msgLen], nil
}

func (b *bgpStream) run() {
	scanner := bufio.NewScanner(&b.r)
	scanner.Split(b.split)

	for scanner.Scan() {
		msg, _ := bgp.ParseBGPMessage(scanner.Bytes())
		if msg.Header.Type == bgp.BGP_MSG_UPDATE {
			b.updates <- msg.Body.(*bgp.BGPUpdate)
		}
	}
}
