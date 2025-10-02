package orchestrator

import (
	"context"
	"fmt"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"connectrpc.com/connect"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/proullon/ramsql/driver"
)

type service struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	db *sqlx.DB
}

// NewService creates a new orchestrator service and returns a [orchestratorconnect.OrchestratorHandler].
func NewService() (orchestratorconnect.OrchestratorHandler, error) {
	var (
		svc = &service{}
		err error
	)

	svc.db, err = sqlx.Open("ramsql", "orchestrator")
	if err != nil {
		return nil, err
	}

	tx := svc.db.MustBegin()
	tx.MustExec("CREATE TABLE target_of_evaluation (id TEXT PRIMARY KEY, name TEXT)")
	tx.MustExec("INSERT INTO target_of_evaluation (id, name) VALUES ($1, $2)", "1", "Hello")
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return svc, nil
}

// ListTargetsOfEvaluation lists all targets of evaluation.
func (svc *service) ListTargetsOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.ListTargetsOfEvaluationRequest],
) (*connect.Response[orchestrator.ListTargetsOfEvaluationResponse], error) {
	var (
		toes = []*orchestrator.TargetOfEvaluation{}
		err  error
	)

	err = svc.db.Select(&toes, "SELECT * FROM target_of_evaluation ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query targets of evaluation: %w", err)
	}

	return connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
		TargetsOfEvaluation: toes,
	}), nil
}
