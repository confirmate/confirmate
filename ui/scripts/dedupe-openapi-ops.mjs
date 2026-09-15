#!/usr/bin/env node
// Walks an OpenAPI YAML file and rewrites duplicate operationIds so each one
// is unique. gnostic-openapi (used upstream to convert the proto annotations
// into OpenAPI) emits the same operationId for every URL binding of an RPC
// when the proto uses google.api.http.additional_bindings. openapi-typescript
// then collapses those into a single TypeScript key and the generated
// declarations stop compiling.
//
// Usage: node scripts/dedupe-openapi-ops.mjs <input.yaml> <output.yaml>
//
// We do the rewrite as a simple line-based pass instead of pulling in a YAML
// dependency: each operationId lives on its own line under a path's HTTP verb,
// and we just append a numeric suffix the second time we see a given id.

import { readFileSync, writeFileSync } from 'node:fs';
import { argv, exit } from 'node:process';

if (argv.length !== 4) {
	console.error('usage: dedupe-openapi-ops.mjs <input.yaml> <output.yaml>');
	exit(2);
}

const [, , inPath, outPath] = argv;
const src = readFileSync(inPath, 'utf8');

const seen = new Map();
const lines = src.split('\n');
for (let i = 0; i < lines.length; i++) {
	const m = lines[i].match(/^(\s*operationId:\s*)(\S+)\s*$/);
	if (!m) continue;
	const prefix = m[1];
	const id = m[2];
	const count = seen.get(id) ?? 0;
	seen.set(id, count + 1);
	if (count === 0) continue;
	lines[i] = `${prefix}${id}_${count + 1}`;
}

writeFileSync(outPath, lines.join('\n'));
