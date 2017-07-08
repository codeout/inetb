package client

import (
	"errors"
	"flag"
	"fmt"
	cli "github.com/osrg/gobgp/client"
	"github.com/osrg/gobgp/config"
	"log"
	"net"
)

var Timeout = flag.Int("t", 10, "Timeout in seconds to wait for convergence")

type Client struct {
	Host          string
	Port          string
	Timeout       int
	GobgpClient   *cli.Client
	Updates       chan *BGPUpdate
	neighbor      *config.Neighbor
	routerId      string
	peerInterface string
}

func New(port string) *Client {
	client, err := cli.New(net.JoinHostPort("127.0.0.1", port))

	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		Port:        port,
		GobgpClient: client,
		Updates:     make(chan *BGPUpdate, 65536),
	}
}

func (c *Client) Init(mrtPath string) error {
	c.Host = "127.0.0.1"

	if err := c.Disable(); err != nil {
		return err
	}

	if err := c.WaitToTurnDown(); err != nil {
		return err
	}

	if err := c.RejectExport(); err != nil {
		return err
	}

	if err := c.AcceptImport(); err != nil {
		return err
	}

	c.LoadRoutes(mrtPath)
	c.Log("Routes have been loaded on %s")

	if err := c.RejectImport(); err != nil {
		return err
	}

	if err := c.Enable(); err != nil {
		return err
	}

	if err := c.WaitToTurnUp(); err != nil {
		return err
	}

	return nil
}

func (c *Client) Log(message string) {
	routerId, _ := c.RouterId()
	log.Printf(message, fmt.Sprintf("router(%s)", routerId))
}

func (c *Client) RouterId() (string, error) {
	if c.routerId != "" {
		return c.routerId, nil
	}

	if server, err := c.GobgpClient.GetServer(); err != nil {
		return "", err
	} else {
		c.routerId = server.Config.RouterId
	}

	return c.routerId, nil
}

func (c *Client) PeerInterface() (string, error) {
	if c.peerInterface != "" {
		return c.peerInterface, nil
	}

	localAddress, err := c.LocalAddress()
	if err != nil {
		return "", err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			switch a := addr.(type) {
			case *net.IPNet:
				if a.IP.String() == localAddress {
					c.peerInterface = iface.Name
					return c.peerInterface, nil
				}
			}
		}
	}

	return "", errors.New(fmt.Sprintf(`No interface associated to "%s"`, localAddress))
}
