package evidence

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service/evidence/evidencetest"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// TestIntegration_Postgres_StoreEvidence_PersistsAndDispatches verifies that
// StoreEvidence works against a real Postgres database (not ramsql).
//
// The test is opt-in and skipped unless CONFIRMATE_IT_POSTGRES=1 is set.
//
// Required environment variables:
// - CONFIRMATE_IT_POSTGRES=1
// - CONFIRMATE_IT_DB_HOST
// - CONFIRMATE_IT_DB_PORT
// - CONFIRMATE_IT_DB_NAME
// - CONFIRMATE_IT_DB_USER
// - CONFIRMATE_IT_DB_PASSWORD
//
// Optional:
// - CONFIRMATE_IT_DB_SSLMODE (default: disable)
func TestIntegration_Postgres_StoreEvidence_PersistsAndDispatches(t *testing.T) {
	var (
		host     string
		portRaw  string
		dbName   string
		user     string
		password string
		sslMode  string
		port     int
		err      error
		recorder *assessmentStreamRecorder
		svc      *Service
		ev       *evidence.Evidence
		storedEv evidence.Evidence
		storedRe evidence.Resource
		res      *connect.Response[evidence.StoreEvidenceResponse]
	)

	if os.Getenv("CONFIRMATE_IT_POSTGRES") != "1" {
		t.Skip("set CONFIRMATE_IT_POSTGRES=1 to run Postgres integration test")
	}

	host = os.Getenv("CONFIRMATE_IT_DB_HOST")
	portRaw = os.Getenv("CONFIRMATE_IT_DB_PORT")
	dbName = os.Getenv("CONFIRMATE_IT_DB_NAME")
	user = os.Getenv("CONFIRMATE_IT_DB_USER")
	password = os.Getenv("CONFIRMATE_IT_DB_PASSWORD")
	sslMode = os.Getenv("CONFIRMATE_IT_DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	if host == "" || portRaw == "" || dbName == "" || user == "" || password == "" {
		t.Skip("missing Postgres integration env vars: CONFIRMATE_IT_DB_HOST, CONFIRMATE_IT_DB_PORT, CONFIRMATE_IT_DB_NAME, CONFIRMATE_IT_DB_USER, CONFIRMATE_IT_DB_PASSWORD")
	}

	port, err = strconv.Atoi(portRaw)
	if err != nil {
		t.Fatalf("invalid CONFIRMATE_IT_DB_PORT %q: %v", portRaw, err)
	}

	recorder, _, srv := newAssessmentTestServer(t)
	defer srv.Close()

	svc, err = NewService(
		WithConfig(Config{
			AssessmentAddress:    srv.URL,
			AssessmentHTTPClient: srv.Client(),
			PersistenceConfig: persistence.Config{
				Host:       host,
				Port:       port,
				DBName:     dbName,
				User:       user,
				Password:   password,
				SSLMode:    sslMode,
				InMemoryDB: false,
				MaxConn:    5,
			},
			EvidenceQueueSize: 64,
		}),
	)
	if !assert.NoError(t, err) {
		t.Fatalf("failed to initialize Postgres-backed evidence service: %v", err)
	}
	defer func() {
		if svc.assessmentStream != nil {
			_ = svc.assessmentStream.Close()
		}
	}()

	ev = proto.Clone(evidencetest.MockEvidenceWithVMResource).(*evidence.Evidence)
	ev.Id = uuid.NewString()
	ev.TargetOfEvaluationId = uuid.NewString()
	ev.ToolId = fmt.Sprintf("it-postgres-%s", uuid.NewString())
	if vm := ev.GetResource().GetVirtualMachine(); vm != nil {
		vm.Id = uuid.NewString()
	}

	res, err = svc.StoreEvidence(context.Background(), connect.NewRequest(&evidence.StoreEvidenceRequest{Evidence: ev}))
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Verify evidence persisted in Postgres.
	err = svc.db.Get(&storedEv, "id = ?", ev.GetId())
	assert.NoError(t, err)
	assert.Equal(t, ev.GetId(), storedEv.Id)
	assert.Equal(t, ev.GetToolId(), storedEv.ToolId)

	// Verify resource persisted in Postgres.
	err = svc.db.Get(&storedRe, "id = ?", ev.GetResource().GetVirtualMachine().GetId())
	assert.NoError(t, err)
	assert.Equal(t, ev.GetTargetOfEvaluationId(), storedRe.TargetOfEvaluationId)
	assert.Equal(t, ev.GetToolId(), storedRe.ToolId)

	// Verify fire-and-forget dispatch to assessment stream is functional.
	awaitAssessmentRequest(t, recorder.received, ev.GetId())
}