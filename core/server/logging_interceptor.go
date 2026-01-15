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
	"log/slog"
	"strings"
	"time"

	"confirmate.io/core/api"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Attribute keys for logging
const (
	// RPC request/response keys
	keyMethod   = "method"
	keyStatus   = "status"
	keyDuration = "duration"
	keyError    = "error"

	// Pagination keys
	keyPageSize      = "page_size"
	keyPageToken     = "page_token"
	keyResults       = "results"
	keyNextPageToken = "next_page_token"

	// Entity keys
	keyId                   = "id"
	keyTargetOfEvaluationId = "target_of_evaluation_id"
	keyToolId               = "tool_id"
	keyPayload              = "payload"

	// Field names to skip in payload
	fieldId        = "id"
	fieldCreatedAt = "created_at"
	fieldUpdatedAt = "updated_at"

	// Enum suffix to skip
	enumUnspecified = "_UNSPECIFIED"

	// ANSI color codes - only used if color is enabled
	ansiReset  = "\033[0m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
)

// LoggingInterceptor logs RPC requests at two levels:
//
//  1. Request-level (INFO/WARN): All requests with method, duration, and status
//  2. Entity-level (DEBUG): Entity operations with details and payloads
type LoggingInterceptor struct{}

// withRequestAttrs extracts common attributes from a request message and stores them in context.
// Supports extracting: id, target_of_evaluation_id, tool_id, and any other *_id fields
func withRequestAttrs(ctx context.Context, msg any) context.Context {
	attrs := make([]slog.Attr, 0, 3)

	if hasId, ok := msg.(api.HasId); ok {
		if id := hasId.GetId(); id != "" {
			attrs = append(attrs, slog.String(keyId, id))
		}
	}

	if hasToeId, ok := msg.(api.HasTargetOfEvaluationId); ok {
		if toeId := hasToeId.GetTargetOfEvaluationId(); toeId != "" {
			attrs = append(attrs, slog.String(keyTargetOfEvaluationId, toeId))
		}
	}

	if hasToolId, ok := msg.(api.HasToolId); ok {
		if toolId := hasToolId.GetToolId(); toolId != "" {
			attrs = append(attrs, slog.String(keyToolId, toolId))
		}
	}

	// Use reflection to extract any other *_id fields (catalog_id, metric_id, etc.)
	if protoMsg, ok := msg.(interface{ ProtoReflect() protoreflect.Message }); ok {
		protoMsg.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			fieldName := string(fd.Name())
			// Check if this is an ID field we haven't already handled
			if strings.HasSuffix(fieldName, "_id") &&
				fieldName != keyId &&
				fieldName != keyTargetOfEvaluationId &&
				fieldName != keyToolId &&
				fd.Kind() == protoreflect.StringKind {
				if strVal := v.String(); strVal != "" {
					attrs = append(attrs, slog.String(fieldName, strVal))
				}
			}
			return true
		})
	}

	if len(attrs) > 0 {
		return log.WithAttrs(ctx, attrs...)
	}
	return ctx
}

// WrapUnary implements the [connect.Interceptor] interface for unary calls.
func (li *LoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (res connect.AnyResponse, err error) {
		var (
			start    = time.Now()
			method   = methodName(req.Spec().Procedure)
			duration time.Duration
		)

		// Add method to context first so it appears before other attributes
		ctx = log.WithAttrs(ctx, slog.String(keyMethod, method))

		// Extract request attributes and store in context for automatic logging
		ctx = withRequestAttrs(ctx, req.Any())

		// Execute the request
		res, err = next(ctx, req)
		duration = time.Since(start)

		// Log combined request and entity information
		li.logRPCRequest(ctx, method, duration, req, res, err)

		return res, err
	}
}

// WrapStreamingClient implements the [connect.Interceptor] interface for streaming client calls.
func (li *LoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next // No streaming logging for now
}

// WrapStreamingHandler implements the [connect.Interceptor] interface for streaming handler calls.
func (li *LoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next // No streaming logging for now
}

// logRPCRequest logs combined request and entity information in a single message.
// CRUD operations (Create/Update/Delete) are logged at INFO level.
// Read operations (Get/List) are logged at DEBUG level.
// Errors are logged at INFO level with color-coded status.
func (li *LoggingInterceptor) logRPCRequest(ctx context.Context, method string, duration time.Duration, req connect.AnyRequest, res connect.AnyResponse, err error) {
	var (
		msg         = req.Any()
		requestType = operationType(method)
		level       = slog.LevelInfo
		status      string
	)

	// Determine log level based on operation type
	if err == nil && isReadOperation(method) {
		level = slog.LevelDebug
	}

	// Build attributes - method is already in context, so start with status
	// Pre-allocate with estimated capacity: status + duration + optional (error/pagination/payload)
	attrs := make([]slog.Attr, 0, 6)

	// Add status with color coding
	if err != nil {
		code := connect.CodeOf(err)
		status = colorCodeStatus(code)

		// Extract just the error message without the code prefix
		errMsg := err.Error()
		if connectErr, ok := err.(*connect.Error); ok {
			errMsg = connectErr.Message()
		}

		attrs = append(attrs,
			slog.String(keyStatus, status),
			slog.Duration(keyDuration, duration),
			slog.String(keyError, errMsg),
		)
	} else {
		if log.ColorEnabled() {
			status = ansiGreen + "ok" + ansiReset
		} else {
			status = "ok"
		}
		attrs = append(attrs,
			slog.String(keyStatus, status),
			slog.Duration(keyDuration, duration),
		)
	}

	// Add pagination info for list operations
	if err == nil && strings.HasPrefix(method, "List") && res != nil {
		if resMsg := res.Any(); resMsg != nil {
			li.addPaginationAttributes(&attrs, msg, resMsg)
		}
	}

	// Add entity payload details for CRUD operations
	if requestType != orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED {
		if payloadReq, ok := msg.(api.PayloadRequest); ok {
			li.addPayloadAttributes(&attrs, requestType, payloadReq, level)
		}
	}

	slog.LogAttrs(ctx, level, "RPC request", attrs...)
}

// addPaginationAttributes adds pagination details for list operations.
func (li *LoggingInterceptor) addPaginationAttributes(attrs *[]slog.Attr, req any, res any) {
	// Extract request pagination info
	if paginatedReq, ok := req.(api.PaginatedRequest); ok {
		if pageSize := paginatedReq.GetPageSize(); pageSize > 0 {
			*attrs = append(*attrs, slog.Int(keyPageSize, int(pageSize)))
		}
		if pageToken := paginatedReq.GetPageToken(); pageToken != "" {
			*attrs = append(*attrs, slog.String(keyPageToken, pageToken))
		}
	}

	// Extract response pagination info
	if paginatedRes, ok := res.(api.PaginatedResponse); ok {
		// Get results count
		if count := api.GetResultsCount(paginatedRes); count > 0 {
			*attrs = append(*attrs, slog.Int(keyResults, count))
		}

		// Get next_page_token
		if nextPageToken := paginatedRes.GetNextPageToken(); nextPageToken != "" {
			*attrs = append(*attrs, slog.String(keyNextPageToken, nextPageToken))
		}
	}
}

// addPayloadAttributes adds entity payload details to the log attributes.
func (li *LoggingInterceptor) addPayloadAttributes(attrs *[]slog.Attr, requestType orchestrator.RequestType, req api.PayloadRequest, level slog.Level) {
	payload := req.GetPayload()
	if payload == nil {
		return
	}

	// Include full payload for write operations (Create/Update) and debug-level reads
	if level == slog.LevelDebug || requestType == orchestrator.RequestType_REQUEST_TYPE_CREATED ||
		requestType == orchestrator.RequestType_REQUEST_TYPE_UPDATED ||
		requestType == orchestrator.RequestType_REQUEST_TYPE_REGISTERED ||
		requestType == orchestrator.RequestType_REQUEST_TYPE_STORED {
		// Use LogValuer wrapper for clean payload logging
		*attrs = append(*attrs, slog.Any(keyPayload, protoPayload{payload}))
	}
}

// protoPayload wraps a proto.Message and implements slog.LogValuer for automatic field extraction.
type protoPayload struct {
	msg protoreflect.ProtoMessage
}

// LogValue implements slog.LogValuer to extract payload fields for logging.
func (p protoPayload) LogValue() slog.Value {
	payloadReflect := p.msg.ProtoReflect()
	// Pre-allocate with estimated capacity based on field count
	attrs := make([]slog.Attr, 0, payloadReflect.Descriptor().Fields().Len())

	payloadReflect.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldName := string(fd.Name())

		// Skip auto-generated fields that weren't in the original request
		if fieldName == fieldId || fieldName == fieldCreatedAt || fieldName == fieldUpdatedAt {
			return true
		}

		// Get string representation - for enums, use the enum name
		var fieldValue string
		if fd.Enum() != nil {
			// It's an enum - get the enum value name
			enumVal := v.Enum()
			enumDesc := fd.Enum().Values().ByNumber(enumVal)
			if enumDesc != nil {
				fieldValue = string(enumDesc.Name())
			}
		} else {
			fieldValue = v.String()
		}

		// Skip fields that are unset, UNSPECIFIED, or empty
		if fieldValue == "" || strings.HasSuffix(fieldValue, enumUnspecified) {
			return true
		}
		attrs = append(attrs, slog.String(fieldName, fieldValue))
		return true
	})

	return slog.GroupValue(attrs...)
}

// colorCodeStatus returns an ANSI color-coded status string for a Connect error code.
// Client errors (4xx equivalent) are colored yellow, server errors (5xx equivalent) are colored red.
// Returns plain code string if colors are disabled.
func colorCodeStatus(code connect.Code) string {
	codeStr := code.String()

	if !log.ColorEnabled() {
		return codeStr
	}

	// Client errors (4xx equivalent) - yellow
	switch code {
	case connect.CodeInvalidArgument,
		connect.CodeFailedPrecondition,
		connect.CodeOutOfRange,
		connect.CodeUnauthenticated,
		connect.CodePermissionDenied,
		connect.CodeNotFound,
		connect.CodeAlreadyExists,
		connect.CodeAborted:
		return ansiYellow + codeStr + ansiReset

	// Server errors (5xx equivalent) - red
	case connect.CodeInternal,
		connect.CodeUnknown,
		connect.CodeDataLoss,
		connect.CodeUnavailable,
		connect.CodeUnimplemented:
		return ansiRed + codeStr + ansiReset

	// Other errors (ResourceExhausted, DeadlineExceeded, Canceled) - red
	default:
		return ansiRed + codeStr + ansiReset
	}
}

// requestTypeToVerb converts an orchestrator.RequestType to a past-tense verb string for logging.
func requestTypeToVerb(rt orchestrator.RequestType) string {
	switch rt {
	case orchestrator.RequestType_REQUEST_TYPE_CREATED:
		return "created"
	case orchestrator.RequestType_REQUEST_TYPE_UPDATED:
		return "updated"
	case orchestrator.RequestType_REQUEST_TYPE_DELETED:
		return "deleted"
	case orchestrator.RequestType_REQUEST_TYPE_REGISTERED:
		return "registered"
	case orchestrator.RequestType_REQUEST_TYPE_STORED:
		return "stored"
	default:
		return "changed"
	}
}

// methodName extracts the method name from a procedure path.
func methodName(procedure string) (name string) {
	if idx := strings.LastIndex(procedure, "/"); idx >= 0 {
		return procedure[idx+1:]
	}
	return procedure
}

// operationType deduces the operation type from a method name.
func operationType(method string) (rt orchestrator.RequestType) {
	switch {
	case strings.HasPrefix(method, "Create"):
		return orchestrator.RequestType_REQUEST_TYPE_CREATED
	case strings.HasPrefix(method, "Update"):
		return orchestrator.RequestType_REQUEST_TYPE_UPDATED
	case strings.HasPrefix(method, "Remove"):
		return orchestrator.RequestType_REQUEST_TYPE_DELETED
	case strings.HasPrefix(method, "Store"):
		return orchestrator.RequestType_REQUEST_TYPE_STORED
	case strings.HasPrefix(method, "Register"):
		return orchestrator.RequestType_REQUEST_TYPE_REGISTERED
	default:
		return orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED
	}
}

// isReadOperation checks if a method represents a read operation.
func isReadOperation(method string) (ok bool) {
	return strings.HasPrefix(method, "Get") ||
		strings.HasPrefix(method, "List") ||
		strings.HasPrefix(method, "Query") ||
		strings.HasPrefix(method, "Search")
}
