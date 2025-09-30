package orchestrator

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"reflect"
	"strconv"

	"confirmate.io/core"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/db"
	"connectrpc.com/connect"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
	_ "github.com/proullon/ramsql/driver"
)

type service struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	db *pgx.Conn
}

func NewService() (orchestratorconnect.OrchestratorHandler, error) {
	var (
		svc = &service{}
		err error
		ctx = context.Background()
	)

	postgres := embeddedpostgres.NewDatabase()
	err = postgres.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start embedded postgres: %w", err)
	}
	defer postgres.Stop()

	// create tables
	if _, err := svc.db.Exec(ctx, core.DDL); err != nil {
		return nil, fmt.Errorf("could not create table: %w", err)
	}

	queries := db.New(svc.db)

	// list all targets of evaluation
	authors, err := queries.ListTargetOfEvaluation(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list targets of evaluation: %w", err)
	}
	log.Println(authors)

	// create a target of evaluation (TOE)
	insertedTOE, err := queries.CreateTargetOfEvaluation(ctx, "TOE1")
	if err != nil {
		return nil, fmt.Errorf("failed to create target of evaluation: %w", err)
	}
	log.Println(insertedTOE)

	// get the TOE we just inserted
	fetchedTOE, err := queries.GetTargetOfEvaluation(ctx, insertedTOE.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target of evaluation: %w", err)
	}

	log.Println(reflect.DeepEqual(insertedTOE, fetchedTOE))

	// tx := svc.db.MustBegin()
	// tx.MustExec("CREATE TABLE target_of_evaluation (id TEXT PRIMARY KEY, name TEXT)")
	// tx.MustExec("INSERT INTO target_of_evaluation (id, name) VALUES ($1, $2)", "1", "Hello")
	// err = tx.Commit()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to commit transaction: %w", err)
	// }

	return svc, nil
}

func (svc *service) ListTargetsOfEvaluation(context.Context, *connect.Request[orchestrator.ListTargetsOfEvaluationRequest]) (*connect.Response[orchestrator.ListTargetsOfEvaluationResponse], error) {
	var (
		toes = []*orchestrator.TargetOfEvaluation{}
		err  error
	)
	queries := db.New(svc.db)

	targetsOfEvaluation, err := queries.ListTargetOfEvaluation(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query targets of evaluation: %w", err)
	}

	for _, v := range targetsOfEvaluation {
		toes = append(toes, &orchestrator.TargetOfEvaluation{
			Id:   strconv.Itoa(int(v.ID)),
			Name: v.Name,
		})
	}

	return connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
		TargetsOfEvaluation: toes,
	}), nil
}
