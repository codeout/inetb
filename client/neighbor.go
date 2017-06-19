package client

import (
	"fmt"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/packet/bgp"
	"time"
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

func (c *Client) Reset() error {
	neighbor, err := c.Neighbor()
	if err != nil {
		return err
	}

	c.Log("Clear neighbor on %s")
	c.GobgpClient.ResetNeighbor(neighbor.Config.NeighborAddress, "Hi!")

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

func (c *Client) WaitToTurnUp() error {
	return c.Wait(config.SESSION_STATE_ESTABLISHED)
}

func (c *Client) WaitToTurnDown() error {
	return c.Wait(config.SESSION_STATE_ESTABLISHED, true)
}

func (c *Client) Wait(state config.SessionState, inverse... bool) error {
	timeout := 180
	var statement string
	var inversed bool

	if len(inverse) > 0 && inverse[0] {
		statement = fmt.Sprintf("not to be %s", state)
		inversed = true
	} else {
		statement = fmt.Sprintf("to be %s", state)
		inversed = false
	}

	c.Log(fmt.Sprintf("Waiting for neighbor %s on %%s", statement))

	for i := 0; i < timeout; i++ {
		neighbor, err := c.Neighbor()
		if err != nil {
			return err
		}

		if inversed {
			if neighbor.State.SessionState != state {
				c.Log(fmt.Sprintf("Neighbor has been %s on %%s", neighbor.State.SessionState))
				return nil
			}
		} else {
			if neighbor.State.SessionState == state {
				c.Log(fmt.Sprintf("Neighbor has been %s on %%s", neighbor.State.SessionState))
				return nil
			}
		}

		time.Sleep(time.Second)
	}

	return nil
}
