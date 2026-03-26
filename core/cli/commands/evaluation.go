package commands

import (
	"context"
	"fmt"
	"strconv"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/util"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

func EvaluationResultsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all evaluation results",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "target",
				Aliases: []string{"t"},
				Usage:   "Filter by target of evaluation ID",
			},
			&cli.StringFlag{
				Name:    "catalog",
				Aliases: []string{"c"},
				Usage:   "Filter by catalog ID",
			},
			&cli.StringFlag{
				Name:    "control",
				Aliases: []string{"co"},
				Usage:   "Filter by control ID",
			},
			&cli.StringFlag{
				Name:    "subcontrol",
				Aliases: []string{"s"},
				Usage:   "Filter by subcontrol ID",
			},
			&cli.BoolFlag{
				Name:    "parent",
				Aliases: []string{"p"},
				Usage:   "Filter for list evaluation results for parents only",
			},
			&cli.BoolFlag{
				Name:    "valid-manual-only",
				Aliases: []string{"v"},
				Usage:   "Filter for valid manual evaluations only",
			},
			&cli.StringFlag{
				Name:    "latest",
				Aliases: []string{"l"},
				Usage:   "Filter by latest evaluation result for each control",
			},
		}, PaginationFlags()...),
		Action: func(ctx context.Context, c *cli.Command) error {
			req := &evaluation.ListEvaluationResultsRequest{
				PageSize:          int32(c.Int("page-size")),
				PageToken:         c.String("page-token"),
				LatestByControlId: util.Ref(c.Bool("latest")),
			}

			// Apply filters if provided
			if c.String("target") != "" || c.String("catalog") != "" || c.String("control") != "" || c.String("subcontrol") != "" || c.IsSet("parent") || c.IsSet("valid-manual-only") {
				filter := &evaluation.ListEvaluationResultsRequest_Filter{}

				if targetID := c.String("target"); targetID != "" {
					filter.TargetOfEvaluationId = &targetID
				}
				if catalogID := c.String("catalog"); catalogID != "" {
					filter.CatalogId = &catalogID
				}
				if controlID := c.String("control"); controlID != "" {
					filter.ControlId = &controlID
				}
				if subcontrolID := c.String("subcontrol"); subcontrolID != "" {
					filter.ControlId = &subcontrolID
				}
				if c.IsSet("parent") {
					parent := c.Bool("parent")
					filter.ParentsOnly = &parent
				}
				if c.IsSet("valid-manual-only") {
					validManualOnly := c.Bool("valid-manual-only")
					filter.ValidManualOnly = &validManualOnly
				}
				req.Filter = filter
			}

			client := EvaluationClient(ctx, c)
			resp, err := client.ListEvaluationResults(ctx, connect.NewRequest(req))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func EvaluationStartCommand() *cli.Command {
	return &cli.Command{
		Name:      "start",
		Usage:     "Start the evaluation of a target",
		ArgsUsage: "<audit-scope-id> <interval>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return fmt.Errorf("audit scope ID and interval are required")
			}
			auditScopeID := c.Args().Get(0)
			interval := c.Args().Get(1)

			client := EvaluationClient(ctx, c)
			resp, err := client.StartEvaluation(ctx, connect.NewRequest(&evaluation.StartEvaluationRequest{
				AuditScopeId: auditScopeID,
				Interval:     util.Ref(toInt32(interval)),
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func EvaluationStopCommand() *cli.Command {
	return &cli.Command{
		Name:      "stop",
		Usage:     "Stop the evaluation of a target",
		ArgsUsage: "<audit-scope-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("audit scope ID is required")
			}
			auditScopeID := c.Args().Get(0)

			client := EvaluationClient(ctx, c)
			resp, err := client.StopEvaluation(ctx, connect.NewRequest(&evaluation.StopEvaluationRequest{
				AuditScopeId: auditScopeID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

// TODO(anatheka): Move to util
func toInt32(s string) int32 {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(n)
}
