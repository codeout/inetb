package client

import (
	"errors"
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
	return c.Wait([]config.SessionState{config.SESSION_STATE_ESTABLISHED})
}

func (c *Client) WaitToTurnDown() error {
	return c.Wait([]config.SessionState{
		config.SESSION_STATE_IDLE,
		config.SESSION_STATE_CONNECT,
		config.SESSION_STATE_ACTIVE,
		config.SESSION_STATE_OPENSENT,
		config.SESSION_STATE_OPENCONFIRM,
	})
}

func (c *Client) Wait(states []config.SessionState) error {
	timeout := 60

	c.Log("Waiting for neighbor to change on %s")

	state, err := func () (config.SessionState, error) {
		for i := 0; i < timeout; i++ {
			neighbor, err := c.Neighbor()
			if err != nil {
				return "", err
			}

			for _, state := range states {
				if neighbor.State.SessionState == state {
					return state, nil
				}
			}

			time.Sleep(time.Second)
		}

		return "", errors.New("Timed out")
	}()
	if err != nil {
		return err
	}

	c.Log(fmt.Sprintf("Neighbor has been %s on %%s", state))
	return nil
}
