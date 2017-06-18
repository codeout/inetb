package client

import (
	"fmt"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/table"
	"log"
	"strconv"
)

func (c *Client) RejectImport() error {
	return c.modifyPolicy(Import, table.ROUTE_TYPE_REJECT)
}

func (c *Client) RejectExport() error {
	return c.modifyPolicy(Export, table.ROUTE_TYPE_REJECT)
}

func (c *Client) AcceptImport() error {
	return c.modifyPolicy(Import, table.ROUTE_TYPE_ACCEPT)
}

func (c *Client) AcceptExport() error {
	return c.modifyPolicy(Export, table.ROUTE_TYPE_ACCEPT)
}

func (c *Client) modifyPolicy(direction Direction, policy table.RouteType) error {
	var assignment *table.PolicyAssignment
	var err error
	var directionText, policyText string

	switch direction {
	case Import:
		assignment, err = c.GobgpClient.GetRouteServerImportPolicy("")
		directionText = "Import"
	case Export:
		assignment, err = c.GobgpClient.GetRouteServerExportPolicy("")
		directionText = "Export"
	}

	if err != nil {
		return err
	}

	switch policy {
	case table.ROUTE_TYPE_REJECT:
		assignment.Default = policy
		policyText = "reject"
	case table.ROUTE_TYPE_ACCEPT:
		assignment.Default = policy
		policyText = "accept"
	}

	if err := c.GobgpClient.ReplacePolicyAssignment(assignment); err != nil {
		return err
	}

	if err := c.SoftReset(); err != nil {
		log.Fatal(err)
	}

	c.Log(fmt.Sprintf("%s policy for %%s is set to \"%s\"", directionText, policyText))
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
				{
					Actions: config.Actions{
						BgpActions: config.BgpActions{
							SetAsPathPrepend: config.SetAsPathPrepend{
								RepeatN: 1,
								As:      strconv.FormatUint(uint64(server.Config.As), 10),
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
				{Name: name},
			},
		}
		if err = c.GobgpClient.ReplacePolicyAssignment(assign); err != nil {
			return err
		}
	}

	c.Log("Depref exporting routes on %s")
	return nil
}
