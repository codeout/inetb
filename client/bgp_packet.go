package client

import (
	"github.com/google/gopacket"
	"github.com/osrg/gobgp/packet/bgp"
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
