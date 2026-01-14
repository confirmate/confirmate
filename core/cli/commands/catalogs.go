package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v3"
	"connectrpc.com/connect"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/api/orchestrator"
)

func CatalogsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all catalogs",
		Action: func(ctx context.Context, c *cli.Command) error {
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.ListCatalogs(ctx, connect.NewRequest(&orchestrator.ListCatalogsRequest{}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func CatalogsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific catalog by ID",
		ArgsUsage: "<catalog-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("catalog ID required")
			}
			catalogID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.GetCatalog(ctx, connect.NewRequest(&orchestrator.GetCatalogRequest{
				CatalogId: catalogID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func CatalogsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete a catalog by ID",
		ArgsUsage: "<catalog-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("catalog ID required")
			}
			catalogID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			_, err := client.RemoveCatalog(ctx, connect.NewRequest(&orchestrator.RemoveCatalogRequest{
				CatalogId: catalogID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("Catalog %s deleted successfully\n", catalogID)
			return nil
		},
	}
}
