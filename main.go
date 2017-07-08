package main

import (
	"flag"
	"fmt"
	"github.com/codeout/inetb/client"
	"github.com/codeout/inetb/test"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime)

	var (
		port1 = flag.String("p1", "50051", "Port number which gobgp1 is listening on")
		port2 = flag.String("p2", "50052", "Port number which gobgp2 is listening on")
		wg    sync.WaitGroup
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MRT_FILE\n\nOptions\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		log.Fatal("MRT Table Dump file is required")
	}

	mrtPath := flag.Arg(0)
	client1 := client.New(*port1)
	client2 := client.New(*port2)

	go client1.StartReader()
	go client2.StartReader()

	time.Sleep(2 * time.Second) // NOTE: Wait for readers

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

	test.AdvertiseNewRoutes(client1, client2)
	test.AdvertiseStrongRoutes(client1, client2)
	test.WithdrawStrongRoutes(client1, client2)
	test.WithdrawLastRoutes(client1, client2)
}
