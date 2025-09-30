package orchestrator

import (
	"context"
	"database/sql"
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

	_ "github.com/lib/pq"
	_ "github.com/proullon/ramsql/driver"
)

type service struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	db *sql.DB
}

func NewService() (orchestratorconnect.OrchestratorHandler, error) {
	var (
		svc = &service{}
		err error
		ctx = context.Background()
	)

	svc.db, err = sql.Open("ramsql", "test")
	if err != nil {
		return nil, err
	}

	// create tables
	if _, err := svc.db.ExecContext(ctx, core.DDL); err != nil {
		return nil, err
	}

	queries := db.New(svc.db)

	// list all targets of evaluation
	authors, err := queries.ListTargetOfEvaluation(ctx)
	if err != nil {
		return nil, err
	}
	log.Println(authors)

	// create an target of evaluation
	insertedTOE, err := queries.CreateTargetOfEvaluation(ctx, "TOE1")
	if err != nil {
		return nil, err
	}
	log.Println(insertedTOE)

	// get the TOE we just inserted
	fetchedTOE, err := queries.GetTargetOfEvaluation(ctx, insertedTOE.ID)
	if err != nil {
		return nil, err
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
