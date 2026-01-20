// Copyright 2016-2025 Fraunhofer AISEC
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

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/log"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

type capturedRecord struct {
	Level   slog.Level
	Message string
	Attrs   []slog.Attr
}

type captureHandler struct {
	mu      sync.Mutex
	records []capturedRecord
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		a.Value = a.Value.Resolve()
		attrs = append(attrs, a)
		return true
	})

	h.mu.Lock()
	h.records = append(h.records, capturedRecord{Level: r.Level, Message: r.Message, Attrs: attrs})
	h.mu.Unlock()
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func (h *captureHandler) lastRecord() (capturedRecord, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.records) == 0 {
		return capturedRecord{}, false
	}
	return h.records[len(h.records)-1], true
}

func stripANSI(s string) string {
	// Best-effort stripping of ANSI escape sequences like "\x1b[31m".
	for {
		start := strings.IndexByte(s, 0x1b)
		if start < 0 {
			return s
		}
		end := strings.IndexByte(s[start:], 'm')
		if end < 0 {
			return s
		}
		s = s[:start] + s[start+end+1:]
	}
}

func groupToMap(v slog.Value) map[string]string {
	v = v.Resolve()
	if v.Kind() != slog.KindGroup {
		return nil
	}
	out := map[string]string{}
	for _, a := range v.Group() {
		a.Value = a.Value.Resolve()
		switch a.Value.Kind() {
		case slog.KindString:
			out[a.Key] = a.Value.String()
		default:
			out[a.Key] = fmt.Sprint(a.Value.Any())
		}
	}
	return out
}

func TestLoggingInterceptor_logRPCRequest(t *testing.T) {
	old := slog.Default()
	defer slog.SetDefault(old)

	h := &captureHandler{}
	slog.SetDefault(slog.New(h))

	li := &LoggingInterceptor{}

	tests := []struct {
		name      string
		ctx       context.Context
		method    string
		duration  time.Duration
		req       connect.AnyRequest
		res       connect.AnyResponse
		err       error
		want      assert.Want[capturedRecord]
		wantFound assert.Want[bool]
	}{
		{
			name:     "create success includes payload",
			ctx:      context.Background(),
			method:   "CreateTargetOfEvaluation",
			duration: 123 * time.Millisecond,
			req:      connect.NewRequest(&orchestrator.CreateTargetOfEvaluationRequest{TargetOfEvaluation: orchestratortest.MockTargetOfEvaluation1}),
			res:      nil,
			err:      nil,
			wantFound: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
			want: func(t *testing.T, got capturedRecord, _ ...any) bool {
				assert.Equal(t, slog.LevelInfo, got.Level)
				assert.Equal(t, "RPC request", got.Message)

				statusAttr, ok := log.FindAttr(got.Attrs, keyStatus)
				assert.True(t, ok)
				assert.Equal(t, "ok", stripANSI(statusAttr.Value.String()))

				durAttr, ok := log.FindAttr(got.Attrs, keyDuration)
				assert.True(t, ok)
				assert.Equal(t, 123*time.Millisecond, durAttr.Value.Duration())

				payloadAttr, ok := log.FindAttr(got.Attrs, keyPayload)
				assert.True(t, ok)
				payload := groupToMap(payloadAttr.Value)
				assert.NotNil(t, payload)
				assert.Equal(t, orchestratortest.MockTargetOfEvaluation1.Name, payload["name"])
				// description may be empty in the shared mock
				_, hasDescription := payload["description"]
				if hasDescription {
					assert.Equal(t, orchestratortest.MockTargetOfEvaluation1.Description, payload["description"])
				}
				assert.Equal(t, "TARGET_TYPE_CLOUD", payload["target_type"])
				_, hasID := payload["id"]
				assert.False(t, hasID)
				return true
			},
		},
		{
			name:     "remove success has no payload",
			ctx:      context.Background(),
			method:   "RemoveTargetOfEvaluation",
			duration: 10 * time.Millisecond,
			req:      connect.NewRequest(&orchestrator.RemoveTargetOfEvaluationRequest{TargetOfEvaluationId: orchestratortest.MockToeID1}),
			res:      nil,
			err:      nil,
			wantFound: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
			want: func(t *testing.T, got capturedRecord, _ ...any) bool {
				_, ok := log.FindAttr(got.Attrs, keyPayload)
				assert.False(t, ok)
				return true
			},
		},
		{
			name:     "list success includes pagination attrs",
			ctx:      context.Background(),
			method:   "ListTargetsOfEvaluation",
			duration: 5 * time.Millisecond,
			req:      connect.NewRequest(&orchestrator.ListTargetsOfEvaluationRequest{PageSize: 25, PageToken: "p1"}),
			res: connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{orchestratortest.MockTargetOfEvaluation1, orchestratortest.MockTargetOfEvaluation2},
				NextPageToken:       "p2",
			}),
			err: nil,
			wantFound: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
			want: func(t *testing.T, got capturedRecord, _ ...any) bool {
				assert.Equal(t, slog.LevelDebug, got.Level)

				pageSizeAttr, ok := log.FindAttr(got.Attrs, keyPageSize)
				assert.True(t, ok)
				assert.Equal(t, int64(25), pageSizeAttr.Value.Int64())

				pageTokenAttr, ok := log.FindAttr(got.Attrs, keyPageToken)
				assert.True(t, ok)
				assert.Equal(t, "p1", pageTokenAttr.Value.String())

				resultsAttr, ok := log.FindAttr(got.Attrs, keyResults)
				assert.True(t, ok)
				assert.Equal(t, int64(2), resultsAttr.Value.Int64())

				nextTokenAttr, ok := log.FindAttr(got.Attrs, keyNextPageToken)
				assert.True(t, ok)
				assert.Equal(t, "p2", nextTokenAttr.Value.String())

				_, ok = log.FindAttr(got.Attrs, keyPayload)
				assert.False(t, ok)
				return true
			},
		},
		{
			name:     "create error includes status and err and payload",
			ctx:      context.Background(),
			method:   "CreateTargetOfEvaluation",
			duration: 1 * time.Millisecond,
			req:      connect.NewRequest(&orchestrator.CreateTargetOfEvaluationRequest{TargetOfEvaluation: orchestratortest.MockTargetOfEvaluation1}),
			res:      nil,
			err:      connect.NewError(connect.CodeInvalidArgument, errors.New("boom")),
			wantFound: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
			want: func(t *testing.T, got capturedRecord, _ ...any) bool {
				assert.Equal(t, slog.LevelInfo, got.Level)

				statusAttr, ok := log.FindAttr(got.Attrs, keyStatus)
				assert.True(t, ok)
				assert.Equal(t, connect.CodeInvalidArgument.String(), stripANSI(statusAttr.Value.String()))

				errAttr, ok := log.FindAttr(got.Attrs, "err")
				assert.True(t, ok)
				assert.Contains(t, fmt.Sprint(errAttr.Value.Any()), "boom")

				payloadAttr, ok := log.FindAttr(got.Attrs, keyPayload)
				assert.True(t, ok)
				payload := groupToMap(payloadAttr.Value)
				assert.Equal(t, orchestratortest.MockTargetOfEvaluation1.Name, payload["name"])
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.mu.Lock()
			h.records = nil
			h.mu.Unlock()

			li.logRPCRequest(tt.ctx, tt.method, tt.duration, tt.req, tt.res, tt.err)

			rec, ok := h.lastRecord()
			tt.wantFound(t, ok)
			tt.want(t, rec)
		})
	}
}
