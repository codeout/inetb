package client

import (
	"github.com/osrg/gobgp/packet/bgp"
)

type Direction int

const (
	Import Direction = iota
	Export
)

type BGPUpdate struct {
	Sequence int
	Nexthop  string
	Raw      *bgp.BGPUpdate
}