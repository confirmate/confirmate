# Agent Instructions

This document holds guidance for automated agents (like Copilot) interacting with the
Confirmate repository.

- For all code changes, read and follow [CONTRIBUTING.md](CONTRIBUTING.md).
  In particular, pay attention to the code style, documentation, and testing
  guidelines there.

- When adding or modifying tests, pay extra attention to the **Testing
  Guidelines** section in [CONTRIBUTING.md](CONTRIBUTING.md).
  The repository contains explicit unit test style rules that are checked by CI,
  so make sure your test code adheres to them.

- If you change authentication or authorization behavior in `core`, update
  [core/docs/authentication-and-authorization.md](core/docs/authentication-and-authorization.md)
  in the same PR.

## UI Development Guidelines

Route files (`+page.svelte`, `+layout.svelte`) should be **minimal** — they receive data via props and pass it to reusable components. Business logic, display logic, and UI patterns live in `$lib/components/`.

- **Route files**: only wire up data to components, handle top-level page structure
- **`$lib/components/`**: all reusable UI — cards, lists, dialogs, forms
  - `nav/` — sidebar, navigation items
  - `ui/` — generic UI primitives (buttons, badges, empty states, etc.)
  - `toe/` — Target of Evaluation specific components

### Tech Stack

- SvelteKit 2 + Svelte 5 (runes: `$props()`, `$state()`, `$derived()`, `{@render children()}`)
- Tailwind CSS v4 (`@tailwindcss/vite`), custom color `--color-confirmate: #005B99`
- `openapi-fetch` with auto-generated types from `openapi-typescript` (`--root-types` flag)
- Pure SPA: `ssr = false` + `trailingSlash = 'always'` in root `+layout.ts`
- Vite proxy: all `/v1/...` requests → `http://localhost:8080`

### API Client Pattern

```ts
import { orchestratorClient } from '$lib/api/client';
// In load functions: orchestratorClient(fetch)
// In browser event handlers: orchestratorClient()
```
