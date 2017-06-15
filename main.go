package main

import (
	"fmt"
	"flag"
	"log"
	"net"
	"sync"
	cli "github.com/osrg/gobgp/client"
)


func newClient(port string) *cli.Client {
	target := net.JoinHostPort("127.0.0.1", port)
	client, err := cli.New(target)

	if err != nil {
		log.Fatal(err)
	}

	return client
}

func serverInfo(client *cli.Client) (string, error) {
	server, err := client.GetServer()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", server.Config.RouterId, server.Config.Port), nil
}

func logWithServer(client *cli.Client, message string) error {
	info, err := serverInfo(client)
	if err != nil {
		return err
	}

	log.Print(fmt.Sprintf(message, info))
	return nil
}

func initServer(client *cli.Client, mrtPath string) error {
	if err := reset(client, true); err != nil {
		return err
	}

	if err := waitForNeighbor(client); err != nil {
		return err
	}

	if err := acceptImport(client); err != nil {
		return err
	}

	if err := rejectExport(client); err != nil {
		return err
	}

	injectMrt(client, mrtPath, -1, 0, true)
	logWithServer(client, "Routes have been loaded on %s")

	if err := rejectImport(client); err != nil {
		return err
	}

	return nil
}


func main() {
	log.SetFlags(log.Ldate | log.Ltime)

	var (
		port1 = flag.String("p1", "50051", "Port number gobgp1 is listening on")
		port2 = flag.String("p2", "50052", "Port number gobgp2 is listening on")
		wg sync.WaitGroup
	)

	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("MRT Table Dump file is required")
	}

	mrtPath := flag.Arg(0)
	client1 := newClient(*port1)
	client2 := newClient(*port2)

	for _, client := range []*cli.Client{client1, client2} {
		wg.Add(1)

		go func(client *cli.Client) {
			if err := initServer(client, mrtPath); err != nil {
				log.Fatal(err)
			}

			wg.Done()
		}(client)
	}
	wg.Wait()

	if err := deprefExport(client1); err != nil {
		log.Fatal(err)
	}
	if err := acceptExport(client1); err != nil {
		log.Fatal(err)
	}

	log.Print("Start benchmarking")
}
