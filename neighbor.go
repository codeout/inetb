package main

import (
	"time"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/packet/bgp"
	cli "github.com/osrg/gobgp/client"
)


func getNeighbor(client *cli.Client) (*config.Neighbor, error) {
	neighbors, err := client.ListNeighbor()
	if err != nil {
		return nil, err
	}

	return neighbors[0], nil
}

func reset(client *cli.Client, hard ...bool) error {
	neighbor, err := getNeighbor(client)
	if err != nil {
		return err
	}

	if len(hard) > 0 && hard[0] {
		logWithServer(client, "Clear neighbor on %s")
		client.ResetNeighbor(neighbor.Config.NeighborAddress, "Hi!")
	} else {
		client.SoftReset(neighbor.Config.NeighborAddress, bgp.RouteFamily(0))
	}

	return nil
}

func softReset(client *cli.Client) error {
	neighbor, err := getNeighbor(client)
	if err != nil {
		return err
	}

	client.SoftReset(neighbor.Config.NeighborAddress, bgp.RouteFamily(0))
	return nil
}

func waitForNeighbor(client *cli.Client) error {
	timeout := 60

	logWithServer(client, "Waiting for neighbor to establish on %s")

	for i := 0; i < timeout; i++ {
		neighbor, err := getNeighbor(client)
		if err != nil {
			return err
		}

		if neighbor.State.SessionState == "established" {
			break
		}

		time.Sleep(time.Second)
	}

	logWithServer(client, "Neighbor has been established on %s")
	return nil
}
