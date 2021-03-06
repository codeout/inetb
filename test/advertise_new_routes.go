package test

import (
	"github.com/codeout/inetb/client"
	"log"
	"time"
)

func AdvertiseNewRoutes(client1 *client.Client, client2 *client.Client) error {
	log.Print("Start benchmarking - Advertise new routes from client1")

	if err := client1.DeprefExport(); err != nil {
		log.Fatal(err)
	}
	if err := client1.AcceptExport(); err != nil {
		log.Fatal(err)
	}

	reports := make([]*Report, 0)
	advertised := 0
	received := 0

	for tick := 0; tick < *client.Timeout; tick++ {
		func() {
			for {
				select {
				case update := <-client1.Updates:
					if client1.IsExportUpdate(update.Net) {
						received += len(update.Raw.NLRI)
					}
					tick = 0
				default:
					return
				}
			}
		}()

		func() {
			for {
				select {
				case update := <-client2.Updates:
					if client2.IsImportUpdate(update.Net) {
						advertised += len(update.Raw.NLRI)
					}
					tick = 0
				default:
					return
				}
			}
		}()

		report := &Report{
			Time:       time.Now().Format("15:04:05"),
			Advertised: advertised,
			Received:   received,
		}

		log.Print(report.String())
		reports = append(reports, report)

		if advertised == 0 && received == 0 {
			tick = 0
		}

		time.Sleep(time.Second)
	}

	log.Print("Stop benchmarking - Advertise new routes from client1")

	return WriteReport("advertise_new_routes.json", reports)
}
