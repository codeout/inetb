package main

import (
	"log"
	"strconv"
	cli "github.com/osrg/gobgp/client"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/table"
)


func rejectImport(client *cli.Client) error {
	assignment, err := client.GetRouteServerImportPolicy("")
	if err != nil {
		return err
	}

	assignment.Default = table.ROUTE_TYPE_REJECT
	if err = client.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := softReset(client); err != nil {
		log.Fatal(err)
	}

	logWithServer(client, "Import policy for %s is set to \"reject\"")
	return nil
}

func rejectExport(client *cli.Client) error {
	assignment, err := client.GetRouteServerExportPolicy("")
	if err != nil {
		return err
	}

	assignment.Default = table.ROUTE_TYPE_REJECT
	if err = client.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := softReset(client); err != nil {
		log.Fatal(err)
	}

	logWithServer(client, "Export policy for %s is set to \"reject\"")
	return nil
}

func acceptImport(client *cli.Client) error {
	assignment, err := client.GetRouteServerImportPolicy("")
	if err != nil {
		return err
	}

	assignment.Default = table.ROUTE_TYPE_ACCEPT
	if err = client.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := softReset(client); err != nil {
		log.Fatal(err)
	}

	logWithServer(client, "Import policy for %s is set to \"accept\"")
	return nil
}

func acceptExport(client *cli.Client) error {
	assignment, err := client.GetRouteServerExportPolicy("")
	if err != nil {
		return err
	}

	assignment.Default = table.ROUTE_TYPE_ACCEPT
	if err = client.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := softReset(client); err != nil {
		log.Fatal(err)
	}

	logWithServer(client, "Export policy for %s is set to \"accept\"")
	return nil
}

func deprefExport(client *cli.Client) error {
	name := "depref"

	policies, err := client.GetPolicy()
	if err != nil {
		return err
	}

	found := false
	for _, p := range policies {
		if p.Name == name {
			found = true
		}
	}

	if !found {
		server, err := client.GetServer()
		if err != nil {
			return err
		}
		
		policy, err := table.NewPolicy(config.PolicyDefinition{
			Name: name,
			Statements: []config.Statement{
				config.Statement{
					Actions: config.Actions{
						BgpActions: config.BgpActions{
							SetAsPathPrepend: config.SetAsPathPrepend{
								RepeatN: 1,
								As: strconv.FormatUint(uint64(server.Config.As), 10),
							},
							SetNextHop: "self",
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}

		client.AddPolicy(policy, true)

		assign := &table.PolicyAssignment{
			Name: "",
			Type: table.POLICY_DIRECTION_EXPORT,
			Policies: []*table.Policy{
				&table.Policy{Name:name},
			},
		}
		if err = client.ReplacePolicyAssignment(assign); err != nil {
			return err
		}
	}

	logWithServer(client, "Depref exporting routes on %s")
	return nil
}
