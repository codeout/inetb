package client

import (
	"github.com/osrg/gobgp/packet/bgp"
	"github.com/google/gopacket"
)

type Direction int

const (
	Import Direction = iota
	Export
)

type BGPUpdate struct {
	Sequence int
	Raw      *bgp.BGPUpdate
	Net      gopacket.Flow
}
