package main

import (
	"fmt"
	"flag"
	"log"
	"sync"
	"time"

	"./client"
	"github.com/osrg/gobgp/packet/bgp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)


const (
	Import = 1
	Export = 2
)

type Report struct {
	Time     string `json:"time"`
	Sent     int    `json:"sent"`
	Received int    `json:"received"`
}

func (r Report) String() string {
	return fmt.Sprintf("%s, Sent: %d, Received: %d", r.Time, r.Sent, r.Received)
}


func readBGPUpdate(client *client.Client, direction int, ch chan *bgp.BGPUpdate) error {
	neighbor, err := client.Neighbor()
	if err != nil {
		return err
	}
	iface, err := client.PeerInterface()
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

func main() {
	log.SetFlags(log.Ldate | log.Ltime)

	var (
		port1 = flag.String("p1", "50051", "Port number which gobgp1 is listening on")
		port2 = flag.String("p2", "50052", "Port number which gobgp2 is listening on")
		wg sync.WaitGroup
	)

	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("MRT Table Dump file is required")
	}

	mrtPath := flag.Arg(0)
	client1 := client.New(*port1)
	client2 := client.New(*port2)

	for _, c := range []*client.Client{client1, client2} {
		wg.Add(1)

		go func(c *client.Client) {
			if err := c.Init(mrtPath); err != nil {
				log.Fatal(err)
			}

			wg.Done()
		}(c)
	}
	wg.Wait()

	exportCh := make(chan *bgp.BGPUpdate)
	importCh := make(chan *bgp.BGPUpdate)
	go readBGPUpdate(client1, Export, exportCh)
	go readBGPUpdate(client2, Import, importCh)


	log.Print("Start benchmarking - Send BGP Update from client1")

	if err := client1.DeprefExport(); err != nil {
		log.Fatal(err)
	}
	if err := client1.AcceptExport(); err != nil {
		log.Fatal(err)
	}

	reports := make([]*Report, 600)
	sent := 0
	received := 0
	timeout := 5

	for tick:=0; tick < timeout; tick++ {
		func() {
			for {
				select {
				case bgp := <- exportCh:
					sent += len(bgp.NLRI)
					tick = 0
				default:
					return
				}
			}
		}()

		func() {
			for {
				select {
				case bgp := <- importCh:
					received += len(bgp.NLRI)
					tick = 0
				default:
					return
				}
			}
		}()

		report := &Report{
			Time: time.Now().Format("15:04:05"),
			Sent: sent,
			Received: received,
		}

		log.Print(report.String())
		reports = append(reports, report)
		time.Sleep(time.Second)
	}

	log.Print("Stop benchmarking - Send BGP Update from client1")
}
