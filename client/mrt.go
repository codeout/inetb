package client

import (
	"flag"
	"fmt"
	"github.com/osrg/gobgp/packet/mrt"
	"github.com/osrg/gobgp/table"
	"io"
	"log"
	"os"
	"time"
)

var prefixes = flag.Int("c", -1, "Max prefix count to be injected. -1 means all of the MRT_FILE")

func (c *Client) LoadRoutes(filename string) {
	if err := c.injectMrt(filename, *prefixes, 0, true); err != nil {
		log.Fatal(err)
	}
}

func (c *Client) injectMrt(filename string, count int, skip int, onlyBest bool) error {
	// Hack
	client := c.GobgpClient
	globalOpts := struct {
		Debug bool
	}{}

	printError := func(err error) {
		log.Print(err)
	}

	exitWithError := func(err error) {
		log.Fatal(err)
	}

	// Ported from github.com/osrg/gobgp/gobgp/cmd/mrt.go
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %s", err)
	}

	idx := 0

	ch := make(chan []*table.Path, 1<<20)

	go func() {

		var peers []*mrt.Peer

		for {
			buf := make([]byte, mrt.MRT_COMMON_HEADER_LEN)
			_, err := file.Read(buf)
			if err == io.EOF {
				break
			} else if err != nil {
				exitWithError(fmt.Errorf("failed to read: %s", err))
			}

			h := &mrt.MRTHeader{}
			err = h.DecodeFromBytes(buf)
			if err != nil {
				exitWithError(fmt.Errorf("failed to parse"))
			}

			buf = make([]byte, h.Len)
			_, err = file.Read(buf)
			if err != nil {
				exitWithError(fmt.Errorf("failed to read"))
			}

			msg, err := mrt.ParseMRTBody(h, buf)
			if err != nil {
				printError(fmt.Errorf("failed to parse: %s", err))
				continue
			}

			if globalOpts.Debug {
				fmt.Println(msg)
			}

			if msg.Header.Type == mrt.TABLE_DUMPv2 {
				subType := mrt.MRTSubTypeTableDumpv2(msg.Header.SubType)
				switch subType {
				case mrt.PEER_INDEX_TABLE:
					peers = msg.Body.(*mrt.PeerIndexTable).Peers
					continue
				case mrt.RIB_IPV4_UNICAST, mrt.RIB_IPV6_UNICAST:
				case mrt.GEO_PEER_TABLE:
					fmt.Printf("WARNING: Skipping GEO_PEER_TABLE: %s", msg.Body.(*mrt.GeoPeerTable))
				default:
					exitWithError(fmt.Errorf("unsupported subType: %v", subType))
				}

				if peers == nil {
					exitWithError(fmt.Errorf("not found PEER_INDEX_TABLE"))
				}

				rib := msg.Body.(*mrt.Rib)
				nlri := rib.Prefix

				paths := make([]*table.Path, 0, len(rib.Entries))

				for _, e := range rib.Entries {
					if len(peers) < int(e.PeerIndex) {
						exitWithError(fmt.Errorf("invalid peer index: %d (PEER_INDEX_TABLE has only %d peers)\n", e.PeerIndex, len(peers)))
					}
					source := &table.PeerInfo{
						AS: peers[e.PeerIndex].AS,
						ID: peers[e.PeerIndex].BgpId,
					}
					t := time.Unix(int64(e.OriginatedTime), 0)
					paths = append(paths, table.NewPath(source, nlri, false, e.PathAttributes, t, false))
				}

				if onlyBest {
					dst := table.NewDestination(nlri, 0)
					for _, p := range paths {
						dst.AddNewPath(p)
					}
					best, _, _ := dst.Calculate().GetChanges(table.GLOBAL_RIB_NAME, false)
					if best == nil {
						exitWithError(fmt.Errorf("Can't find the best %v", nlri))
					}
					paths = []*table.Path{best}
				}

				if idx >= skip {
					ch <- paths
				}

				idx += 1
				if idx == count+skip {
					break
				}
			}
		}

		close(ch)
	}()

	stream, err := client.AddPathByStream()
	if err != nil {
		return fmt.Errorf("failed to modpath: %s", err)
	}

	for paths := range ch {
		err = stream.Send(paths...)
		if err != nil {
			return fmt.Errorf("failed to send: %s", err)
		}
	}

	if err := stream.Close(); err != nil {
		return fmt.Errorf("failed to send: %s", err)
	}
	return nil
}
