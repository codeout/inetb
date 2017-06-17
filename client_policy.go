package main

import (
	"log"
	"strconv"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/table"
)


func (c *Client) RejectImport() error {
	assignment, err := c.GobgpClient.GetRouteServerImportPolicy("")
	if err != nil {
		return err
	}

	assignment.Default = table.ROUTE_TYPE_REJECT
	if err = c.GobgpClient.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := c.SoftReset(); err != nil {
		log.Fatal(err)
	}

	c.Log("Import policy for %s is set to \"reject\"")
	return nil
}

func (c *Client) RejectExport() error {
	assignment, err := c.GobgpClient.GetRouteServerExportPolicy("")
	if err != nil {
		return err
	}

	assignment.Default = table.ROUTE_TYPE_REJECT
	if err = c.GobgpClient.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := c.SoftReset(); err != nil {
		log.Fatal(err)
	}

	c.Log("Export policy for %s is set to \"reject\"")
	return nil
}

func (c *Client) AcceptImport() error {
	assignment, err := c.GobgpClient.GetRouteServerImportPolicy("")
	if err != nil {
		return err
	}

	assignment.Default = table.ROUTE_TYPE_ACCEPT
	if err = c.GobgpClient.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := c.SoftReset(); err != nil {
		log.Fatal(err)
	}

	c.Log("Import policy for %s is set to \"accept\"")
	return nil
}

func (c *Client) AcceptExport() error {
	assignment, err := c.GobgpClient.GetRouteServerExportPolicy("")
	if err != nil {
		return err
	}

	assignment.Default = table.ROUTE_TYPE_ACCEPT
	if err = c.GobgpClient.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := c.SoftReset(); err != nil {
		log.Fatal(err)
	}

	c.Log("Export policy for %s is set to \"accept\"")
	return nil
}

func (c *Client) DeprefExport() error {
	name := "depref"

	policies, err := c.GobgpClient.GetPolicy()
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
		server, err := c.GobgpClient.GetServer()
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

		c.GobgpClient.AddPolicy(policy, true)

		assign := &table.PolicyAssignment{
			Name: "",
			Type: table.POLICY_DIRECTION_EXPORT,
			Policies: []*table.Policy{
				&table.Policy{Name:name},
			},
		}
		if err = c.GobgpClient.ReplacePolicyAssignment(assign); err != nil {
			return err
		}
	}

	c.Log("Depref exporting routes on %s")
	return nil
}
