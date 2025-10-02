package core

//go:generate buf generate
//go:generate buf generate --template buf.openapi.gen.yaml --path api/orchestrator -o openapi/orchestrator
//go:generate buf generate --template buf.gen.ontology.yaml --path policies/security-metrics/ontology/1.0/ontology.proto -o api/ontology
