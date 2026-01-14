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

func CertificatesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all certificates",
		Action: func(ctx context.Context, c *cli.Command) error {
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.ListCertificates(ctx, connect.NewRequest(&orchestrator.ListCertificatesRequest{}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func CertificatesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific certificate by ID",
		ArgsUsage: "<certificate-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("certificate ID required")
			}
			certID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.GetCertificate(ctx, connect.NewRequest(&orchestrator.GetCertificateRequest{
				CertificateId: certID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func CertificatesDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete a certificate by ID",
		ArgsUsage: "<certificate-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("certificate ID required")
			}
			certID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			_, err := client.RemoveCertificate(ctx, connect.NewRequest(&orchestrator.RemoveCertificateRequest{
				CertificateId: certID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("Certificate %s deleted successfully\n", certID)
			return nil
		},
	}
}
