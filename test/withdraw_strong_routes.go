package test

import (
	"github.com/codeout/inetb/client"
	"log"
	"time"
)

func WithdrawStrongRoutes(client1 *client.Client, client2 *client.Client) error {
	log.Print("Start benchmarking - Withdraw strong routes from client2")

	if err := client2.RejectExport(); err != nil {
		log.Fatal(err)
	}

	reports := make([]*Report, 0)
	advertised := 0
	received := 0

	for tick := 0; tick < *client.Timeout; tick++ {
		func() {
			for {
				select {
				case update := <-client2.Updates:
					if client2.IsExportUpdate(update.Net) {
						received += len(update.Raw.WithdrawnRoutes)
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
				case update := <-client1.Updates:
					if client1.IsImportUpdate(update.Net) {
						// NOTE: Some implementations advertise NLRI instead of withdrawn routes,
						//       so count both of them.
						//       Somehow they advertise NLRI back to the peer which they learned
						//       the NLRI from, then they will withdraw only when the last route
						//       disappears from RIB. Otherwise they will advertise NLRI.
						advertised += len(update.Raw.WithdrawnRoutes) + len(update.Raw.NLRI)
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

	log.Print("Stop benchmarking - Withdraw strong routes from client2")

	return WriteReport("withdraw_strong_routes.json", reports)
}
