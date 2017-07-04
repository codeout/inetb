package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/codeout/inetb/client"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type Report struct {
	Time     string `json:"time"`
	Sent     int    `json:"sent"`
	Received int    `json:"received"`
}

func (r Report) String() string {
	return fmt.Sprintf("%s, Sent: %d, Received: %d", r.Time, r.Sent, r.Received)
}

func WriteReport(file string, reports []*Report) error {
	json, err := json.Marshal(reports)
	if err != nil {
		return err
	}

	dir := "reports"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path.Join(dir, file), json, 0644)
}

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

	advertiseNewRoutes(client1, client2)
	advertiseStrongRoutes(client1, client2)
	withdrawStrongRoutes(client1, client2)
	withdrawLastRoutes(client1, client2)
}
