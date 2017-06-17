package client

import (
	"time"
	"github.com/osrg/gobgp/packet/bgp"
	"github.com/osrg/gobgp/config"
)


func (c *Client) Neighbor() (*config.Neighbor, error) {
	if c.neighbor != nil {
		return c.neighbor, nil
	}

	if neighbors, err := c.GobgpClient.ListNeighbor(); err != nil {
		return nil, err
	} else {
		c.neighbor = neighbors[0]
	}

	return c.neighbor, nil
}

func (c *Client) Reset(hard ...bool) error {
	neighbor, err := c.Neighbor()
	if err != nil {
		return err
	}

	if len(hard) > 0 && hard[0] {
		c.Log("Clear neighbor on %s")
		c.GobgpClient.ResetNeighbor(neighbor.Config.NeighborAddress, "Hi!")
	} else {
		c.GobgpClient.SoftReset(neighbor.Config.NeighborAddress, bgp.RouteFamily(0))
	}

	return nil
}

func (c *Client) SoftReset() error {
	neighbor, err := c.Neighbor()
	if err != nil {
		return err
	}

	c.GobgpClient.SoftReset(neighbor.Config.NeighborAddress, bgp.RouteFamily(0))
	return nil
}

func (c *Client) WaitToEstablish() error {
	timeout := 60

	c.Log("Waiting for neighbor to establish on %s")

	for i := 0; i < timeout; i++ {
		neighbor, err := c.Neighbor()
		if err != nil {
			return err
		}

		if neighbor.State.SessionState == "established" {
			break
		}

		time.Sleep(time.Second)
	}

	c.Log("Neighbor has been established on %s")
	return nil
}
