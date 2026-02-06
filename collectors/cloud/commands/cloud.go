package commands

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	cloud "confirmate.io/collectors/cloud/service"
	"confirmate.io/core/service"
	"github.com/urfave/cli/v3"
)

var CloudCollectorCommand = &cli.Command{
	Name:  "cloud-collector",
	Usage: "Launches the cloud collector service",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:     "collector-provider",
			Aliases:  []string{"p"},
			Usage:    "Cloud Provider (aws, azure, openstack, k8s, csaf, openstack)",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "target-of-evaluation-id",
			Aliases:  []string{"e"},
			Usage:    "Target of evaluation ID for which to collect the cloud evidence",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "collector-tool-id",
			Aliases:  []string{"t"},
			Usage:    "Collector Tool ID to identify the collector instance",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "collector-collector",
			Aliases:  []string{"c"},
			Usage:    "Additional collector to use.",
			Required: false,
		},
		&cli.IntFlag{
			Name:     "collector-interval",
			Aliases:  []string{"i"},
			Usage:    "Interval in minutes for periodic collection. (Default: 5 minutes)",
			Required: false,
		},
		&cli.BoolFlag{
			Name:     "collector-auto-start",
			Aliases:  []string{"a"},
			Usage:    "Collector starts automatically after launching the service. (Default: false)",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "collector-resource-group",
			Aliases:  []string{"r"},
			Usage:    "Limit the scope of the collector to a specific resource group.",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "collector-csaf-domain",
			Aliases:  []string{"d"},
			Usage:    "CSAF domain to fetch the CSAF documents from.",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "collector-evidence-store-address",
			Aliases:  []string{"e"},
			Usage:    "Address of the evidence store to send collected evidence to. (default: localhost:9092)",
			Required: false,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var (
			svc  *cloud.Service
			opts []service.Option[*cloud.Service]
		)

		if len(cmd.StringSlice("collector-provider")) > 0 {
			opts = append(opts,
				cloud.WithProviders(cmd.StringSlice("collector-provider")))
		}
		if cmd.String("target-of-evaluation-id") != "" {
			opts = append(opts, cloud.WithTargetOfEvaluationID(cmd.String("target-of-evaluation-id")))
		}
		if cmd.String("collector-tool-id") != "" {
			opts = append(opts, cloud.WithCollectorToolID(cmd.String("collector-tool-id")))
		}
		// if cmd.String("collector-collector") != "" {
		// 	opts = append(opts, cloud.WithAdditionalCollectors(cmd.String("collector-collector")))
		// }
		if cmd.Int("collector-interval") != 0 {
			opts = append(opts, cloud.WithCollectorInterval(cmd.Duration("collector-interval")))
		}
		if cmd.String("collector-evidence-store-address") != "" {
			opts = append(opts, cloud.WithEvidenceStoreAddress("collector-evidence-store-address", http.DefaultClient))
		}

		svc = cloud.NewService(opts...)
		svc.Init(ctx, cmd)

		// Signal-Context (blocks bis SIGINT/SIGTERM)
		sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		<-sigCtx.Done() // Wait until signal

		return nil
	},
}
