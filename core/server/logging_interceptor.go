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
			attrs = append(attrs, slog.String("id", id))
		}
	}

	if hasToeId, ok := msg.(api.HasTargetOfEvaluationId); ok {
		if toeId := hasToeId.GetTargetOfEvaluationId(); toeId != "" {
			attrs = append(attrs, slog.String("target_of_evaluation_id", toeId))
		}
	}

	if hasToolId, ok := msg.(api.HasToolId); ok {
		if toolId := hasToolId.GetToolId(); toolId != "" {
			attrs = append(attrs, slog.String("tool_id", toolId))
		}
	}

	// Use reflection to extract any other *_id fields (catalog_id, metric_id, etc.)
	if protoMsg, ok := msg.(interface{ ProtoReflect() protoreflect.Message }); ok {
		protoMsg.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			fieldName := string(fd.Name())
			// Check if this is an ID field we haven't already handled
			if strings.HasSuffix(fieldName, "_id") && 
			   fieldName != "id" && 
			   fieldName != "target_of_evaluation_id" && 
			   fieldName != "tool_id" &&
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

		// Extract request attributes and store in context for automatic logging
		ctx = withRequestAttrs(ctx, req.Any())

		// Execute the request
		res, err = next(ctx, req)
		duration = time.Since(start)

		// Log entity-level details first (both success and failure)
		li.logEntityOperation(ctx, method, req, err)

		// Log request-level summary
		li.logRPCRequest(ctx, method, duration, err)

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

// logRPCRequest logs request-level information (like HTTP access logs).
func (li *LoggingInterceptor) logRPCRequest(ctx context.Context, method string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("method", method),
		slog.Duration("duration", duration),
	}

	if err != nil {
		// Extract just the error message without the code prefix
		errMsg := err.Error()
		if connectErr, ok := err.(*connect.Error); ok {
			errMsg = connectErr.Message()
		}
		
		// Add error code first, then message (context attributes like catalog_id will follow)
		attrs = append(attrs,
			slog.String("code", connect.CodeOf(err).String()),
			slog.String("error", errMsg),
		)
		slog.LogAttrs(ctx, slog.LevelWarn, "Request failed", attrs...)
	} else {
		slog.LogAttrs(ctx, slog.LevelInfo, "Request completed", attrs...)
	}
}

// logEntityOperation logs entity-level operations at debug level.
func (li *LoggingInterceptor) logEntityOperation(ctx context.Context, method string, req connect.AnyRequest, err error) {
	var (
		msg         = req.Any()
		requestType = operationType(method)
	)

	// Handle write operations (Create/Update/Delete/Store/etc)
	if requestType != orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED {
		if payloadReq, ok := msg.(api.PayloadRequest); ok {
			// Include error info if operation failed
			if err != nil {
				li.logRequest(ctx, slog.LevelDebug, requestType, payloadReq,
					slog.String("error", err.Error()),
					slog.String("code", connect.CodeOf(err).String()),
				)
			} else {
				li.logRequest(ctx, slog.LevelDebug, requestType, payloadReq)
			}
			return
		}
	}

	// Handle read operations (Get/List/Query/Search) - only log successful reads
	if err == nil && isReadOperation(method) {
		li.logReadAccess(ctx, method, msg)
	}
}

// logReadAccess logs read operations.
// Request attributes (id, target_of_evaluation_id, tool_id) are automatically included from context.
func (li *LoggingInterceptor) logReadAccess(ctx context.Context, method string, msg any) {
	// Extract entity name from method (e.g., "GetCatalog" → "Catalog", "ListCatalogs" → "Catalogs")
	entityType := method
	if strings.HasPrefix(method, "Get") {
		entityType = strings.TrimPrefix(method, "Get")
	} else if strings.HasPrefix(method, "List") {
		entityType = strings.TrimPrefix(method, "List")
	} else if strings.HasPrefix(method, "Query") {
		entityType = strings.TrimPrefix(method, "Query")
	} else if strings.HasPrefix(method, "Search") {
		entityType = strings.TrimPrefix(method, "Search")
	}
	
	slog.DebugContext(ctx, entityType+" accessed")
}

// logRequest logs entity operations with full payload details at DEBUG level.
func (li *LoggingInterceptor) logRequest(ctx context.Context, level slog.Level, requestType orchestrator.RequestType, req api.PayloadRequest, attrs ...slog.Attr) {
	// Get the payload from the request
	payload := req.GetPayload()

	// Use the payload type name (e.g., "Catalog") instead of request type name (e.g., "CreateCatalogRequest")
	var name string
	if payload != nil {
		name = string(payload.ProtoReflect().Descriptor().Name())
	} else {
		// Fallback to request type name if no payload
		name = string(req.ProtoReflect().Descriptor().Name())
	}

	// Build structured log attributes
	logAttrs := make([]slog.Attr, 0, 2+len(attrs))

	// Extract ID from payload if available and add as top-level attribute
	if payload != nil {
		if hasId, ok := payload.(api.HasId); ok {
			if id := hasId.GetId(); id != "" {
				logAttrs = append(logAttrs, slog.String("id", id))
			}
		}
	}

	// For debug level, include the full payload for detailed inspection
	if level == slog.LevelDebug && payload != nil {
		// Find the field name that contains the payload in the request
		// (e.g., "catalog" in CreateCatalogRequest)
		payloadFieldName := "payload" // default
		reqReflect := req.ProtoReflect()
		reqReflect.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			// Check if this field's value is the payload
			if fd.Message() != nil && v.Message().Interface() == payload {
				payloadFieldName = string(fd.Name())
				return false // stop iteration
			}
			return true
		})

		// Extract payload fields into a slog.Group using the actual field name
		payloadAttrs := make([]any, 0)
		payloadReflect := payload.ProtoReflect()
		payloadReflect.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			payloadAttrs = append(payloadAttrs, slog.String(string(fd.Name()), v.String()))
			return true
		})
		logAttrs = append(logAttrs, slog.Group(payloadFieldName, payloadAttrs...))
	}

	// Add any additional attributes provided by the caller
	logAttrs = append(logAttrs, attrs...)

	// Build simple message: "PayloadType verb" (e.g., "Catalog created")
	verb := requestTypeToVerb(requestType)
	msg := name + " " + verb

	// Log the message with structured attributes (context attributes added automatically)
	slog.LogAttrs(ctx, level, msg, logAttrs...)
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
