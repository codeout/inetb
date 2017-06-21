package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"
	"github.com/codeout/inetb/client"
)

type Report struct {
	Time     string `json:"time"`
	Sent     int    `json:"sent"`
	Received int    `json:"received"`
}

func (r Report) String() string {
	return fmt.Sprintf("%s, Sent: %d, Received: %d", r.Time, r.Sent, r.Received)
}

func advertiseNewRoutes(client1 *client.Client, client2 *client.Client) {
	log.Print("Start benchmarking - Advertise new routes from client1")

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

	for tick := 0; tick < timeout; tick++ {
		func() {
			for {
				select {
				case update := <- client1.Updates:
					sent += len(update.Raw.NLRI)
					tick = 0
				default:
					return
				}
			}
		}()

		func() {
			for {
				select {
				case update := <- client2.Updates:
					received += len(update.Raw.NLRI)
					tick = 0
				default:
					return
				}
			}
		}()

		report := &Report{
			Time:     time.Now().Format("15:04:05"),
			Sent:     sent,
			Received: received,
		}

		log.Print(report.String())
		reports = append(reports, report)
		time.Sleep(time.Second)
	}

	log.Print("Stop benchmarking - Advertise new routes from client1")
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime)

	var (
		port1 = flag.String("p1", "50051", "Port number which gobgp1 is listening on")
		port2 = flag.String("p2", "50052", "Port number which gobgp2 is listening on")
		wg    sync.WaitGroup
	)

	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("MRT Table Dump file is required")
	}

	mrtPath := flag.Arg(0)
	client1 := client.New(*port1)
	client2 := client.New(*port2)

	go client1.StartReader()
	go client2.StartReader()

	time.Sleep(2*time.Second)  // NOTE: Wait for readers

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

	advertiseNewRoutes(client1, client2)
}
