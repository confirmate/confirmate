package core

//go:generate buf generate
//go:generate buf generate --template buf.openapi.gen.yaml --path api/orchestrator -o openapi/orchestrator
//go:generate buf generate --exclude-path="internal/ontology/clouditor_header.proto" --template buf.gotag.gen.yaml
