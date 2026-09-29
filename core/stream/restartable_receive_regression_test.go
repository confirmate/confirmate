// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package stream_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/stream"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

// rejectingInterceptor rejects every streaming request before it reaches the handler,
// simulating a persistent server-side failure (e.g. authentication rejection).
type rejectingInterceptor struct{}

func (rejectingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc { return next }

func (rejectingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (rejectingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(_ context.Context, _ connect.StreamingHandlerConn) error {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth token"))
	}
}

// TestRestartableBidiStream_Receive_ReturnsErrorInsteadOfHanging_AfterRestart is a regression test:
// previously, when a Receive() error triggered a stream restart, Receive() was retried on the
// freshly restarted stream without ever calling Send() on it again. Since nothing was ever sent on
// that stream, the retried Receive() blocked forever waiting for a response that would never
// arrive, deadlocking any caller (e.g. [confirmate.io/core/service/collection.Service]) that hit a
// persistent server-side rejection such as a missing/invalid auth token.
func TestRestartableBidiStream_Receive_ReturnsErrorInsteadOfHanging_AfterRestart(t *testing.T) {
	handler := &evidenceconnect.UnimplementedEvidenceStoreHandler{}
	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(handler, connect.WithInterceptors(rejectingInterceptor{}))),
	)
	defer testSrv.Close()

	client := evidenceconnect.NewEvidenceStoreClient(testSrv.Client(), testSrv.URL)
	factory := func(ctx context.Context) *connect.BidiStreamForClient[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse] {
		return client.StoreEvidences(ctx)
	}

	rs, err := stream.NewRestartableBidiStream(context.Background(), factory, stream.DefaultRestartConfig(), "TestStream")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, rs.Close())
	}()

	assert.NoError(t, rs.Send(&evidence.StoreEvidenceRequest{}))

	done := make(chan error, 1)
	go func() {
		_, receiveErr := rs.Receive()
		done <- receiveErr
	}()

	select {
	case receiveErr := <-done:
		assert.Error(t, receiveErr)
		assert.ErrorContains(t, receiveErr, "resend required")
		assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(receiveErr))
	case <-time.After(5 * time.Second):
		t.Fatal("Receive() did not return promptly after a restart; the stream deadlocked")
	}
}
