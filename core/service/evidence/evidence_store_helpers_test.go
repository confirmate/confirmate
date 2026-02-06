package evidence

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

// assessmentStreamRecorder captures requests sent to the assessment stream.
type assessmentStreamRecorder struct {
	assessmentconnect.UnimplementedAssessmentHandler
	received chan *assessment.AssessEvidenceRequest
}

// AssessEvidences implements the streaming RPC and records received requests.
func (r *assessmentStreamRecorder) AssessEvidences(
	_ context.Context,
	stream *connect.BidiStream[assessment.AssessEvidenceRequest, assessment.AssessEvidencesResponse],
) error {
	for {
		msg, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		r.received <- msg
		if err = stream.Send(&assessment.AssessEvidencesResponse{
			Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
		}); err != nil {
			return err
		}
	}
}

// newAssessmentTestServer spins up a TLS-enabled in-memory assessment server for stream tests.
func newAssessmentTestServer(t *testing.T) (*assessmentStreamRecorder, *server.Server, *httptest.Server) {
	t.Helper()
	recorder := &assessmentStreamRecorder{
		received: make(chan *assessment.AssessEvidenceRequest, 10),
	}
	srv, testSrv := servertest.NewTestConnectServer(
		t,
		server.WithHandler(assessmentconnect.NewAssessmentHandler(recorder)),
	)
	return recorder, srv, testSrv
}

// awaitAssessmentRequest blocks until a request arrives or a timeout triggers a test failure.
func awaitAssessmentRequest(t *testing.T, ch <-chan *assessment.AssessEvidenceRequest, wantID string) {
	t.Helper()
	select {
	case msg := <-ch:
		assert.Equal(t, wantID, msg.GetEvidence().GetId())
	case <-time.After(2 * time.Second):
		assert.Fail(t, "timed out waiting for assessment request")
	}
}

// nilAssessmentClient is used to force stream factory failures in tests.
type nilAssessmentClient struct{}

func (nilAssessmentClient) CalculateCompliance(context.Context, *connect.Request[assessment.CalculateComplianceRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, errors.New("not implemented")
}

func (nilAssessmentClient) AssessEvidence(context.Context, *connect.Request[assessment.AssessEvidenceRequest]) (*connect.Response[assessment.AssessEvidenceResponse], error) {
	return nil, errors.New("not implemented")
}

func (nilAssessmentClient) AssessEvidences(context.Context) *connect.BidiStreamForClient[assessment.AssessEvidenceRequest, assessment.AssessEvidencesResponse] {
	return nil
}

// fakeReceive describes the next Receive result for a fake stream.
type fakeReceive struct {
	req *evidence.StoreEvidenceRequest
	err error
}

// fakeEvidenceStream simulates a bidi stream with configurable receive/send behavior.
type fakeEvidenceStream struct {
	receives []fakeReceive
	idx      int
	sendErr  error
}

// Receive returns predefined messages and errors in order.
func (f *fakeEvidenceStream) Receive() (*evidence.StoreEvidenceRequest, error) {
	if f.idx >= len(f.receives) {
		return nil, io.EOF
	}
	item := f.receives[f.idx]
	f.idx++
	return item.req, item.err
}

// Send returns the configured error to simulate send failures.
func (f *fakeEvidenceStream) Send(*evidence.StoreEvidencesResponse) error {
	return f.sendErr
}
