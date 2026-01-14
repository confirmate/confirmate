package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"connectrpc.com/connect"
	"confirmate.io/core/api/orchestrator"
)

func CatalogsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all catalogs",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(c)
			resp, err := client.ListCatalogs(ctx, connect.NewRequest(&orchestrator.ListCatalogsRequest{
				PageSize:  int32(c.Int("page-size")),
				PageToken: c.String("page-token"),
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func CatalogsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific catalog by ID",
		ArgsUsage: "<catalog-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("catalog ID required")
			}
			catalogID := c.Args().Get(0)
			
			client := OrchestratorClient(c)
			resp, err := client.GetCatalog(ctx, connect.NewRequest(&orchestrator.GetCatalogRequest{
				CatalogId: catalogID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
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
			if c.Args().Len() < 1 {
				return fmt.Errorf("catalog ID required")
			}
			catalogID := c.Args().Get(0)
			
			client := OrchestratorClient(c)
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
