package test

import (
	"github.com/codeout/inetb/client"
	"log"
	"time"
)

func WithdrawLastRoutes(client1 *client.Client, client2 *client.Client) error {
	log.Print("Start benchmarking - Withdraw last routes from client1")

	if err := client1.RejectExport(); err != nil {
		log.Fatal(err)
	}

	reports := make([]*Report, 0)
	sent := 0
	received := 0

	for tick := 0; tick < *client.Timeout; tick++ {
		func() {
			for {
				select {
				case update := <-client1.Updates:
					if client1.IsExportUpdate(update.Net) {
						sent += len(update.Raw.WithdrawnRoutes)
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
						received += len(update.Raw.WithdrawnRoutes)
					}
					tick = 0
				default:
					return
				}
			}
		}()

		report := &Report{
			Time:     time.Now().Format("15:04:05"),
			Sent:     sent,
			Received: received,
		}

		log.Print(report.String())
		reports = append(reports, report)

		if sent == 0 && received == 0 {
			tick = 0
		}

		time.Sleep(time.Second)
	}

	log.Print("Stop benchmarking - Withdraw last routes from client1")

	return WriteReport("withdraw_last_routes.json", reports)
}
