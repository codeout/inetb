package main

import (
	"fmt"
	"log"
	"net"
	cli "github.com/osrg/gobgp/client"
	"github.com/osrg/gobgp/config"
)


type Client struct {
	Host string
	Port string
	GobgpClient *cli.Client
	neighbor *config.Neighbor
	routerId string
}


func newClient(port string) *Client {
	client, err := cli.New(net.JoinHostPort("127.0.0.1", port))

	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		Port: port,
		GobgpClient: client,
	}
}


func (c *Client) Init(mrtPath string) error {
	c.Host = "127.0.0.1"

	if err := reset(c, true); err != nil {
		return err
	}

	if err := c.WaitToEstablish(); err != nil {
		return err
	}

	if err := c.AcceptImport(); err != nil {
		return err
	}

	if err := c.RejectExport(); err != nil {
		return err
	}

	c.LoadRoutes(mrtPath)
	c.Log("Routes have been loaded on %s")

	if err := c.RejectImport(); err != nil {
		return err
	}

	return nil
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

func (c *Client) Log(message string) {
	routerId, _ := c.RouterId()
	log.Printf(message, fmt.Sprintf("router(%s)", routerId))
}
