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

## Code Generation

When modifying proto files or adding new API services, run the following in `core/`:

```bash
cd core
go generate ./...
buf generate
```

Then regenerate frontend types in `ui/`:

```bash
cd ui
npm run generate:types
```

**Important:** After running `go generate`, you must rebuild and restart the demo for changes to take effect.

## UI Development Guidelines

Route files (`+page.svelte`, `+layout.svelte`) should be **minimal** — they receive data via props and pass it to reusable components. Business logic, display logic, and UI patterns live in `$lib/components/`.

- **Route files**: only wire up data to components, handle top-level page structure
- **`$lib/components/`**: all reusable UI — cards, lists, dialogs, forms
  - `nav/` — sidebar, navigation items
  - `ui/` — generic UI primitives (buttons, badges, empty states, etc.)
  - `toe/` — Target of Evaluation specific components

### Guided experience and design layer

The guided experience lives under `ui/src/routes/guided/` and is a preview that is off by default (feature flag
`VITE_FEATURE_GUIDED=true` at build time, or `localStorage.setItem('cf.feature.guided', 'true')` in the browser).
Existing screens do not change. Read [ui/docs/design-guide.md](ui/docs/design-guide.md) before you change it. The rules
that always apply:

- Route files contain no markup of their own: they pass loaded data to one component. Reuse components from
  `ui/src/lib/ds` and `ui/src/lib/guided/components` before you write new markup.
- Use only the semantic design tokens (`bg-surface`, `text-ink`, `border-line`, `bg-brand`, `text-ok`,
  `rounded-card`, ...) defined in `ui/src/lib/theme/theme.css`, never raw palette classes or hex colors.
- Keep `ui/src/lib/ds` free of domain knowledge, and keep framework specifics (CRA, C5, NIS-2) in a profile under
  `ui/src/lib/guided/profiles/`, not in components.
- Read and write data only through `GuidedSource` (`ui/src/lib/guided/select.ts`), which serves the orchestrator
  (`real.ts`) or sample data (`mock.ts`) depending on the Data switch in the top bar; do not call the API from guided
  components. A new operation goes into the interface and into both implementations.
- Use "Target of Evaluation" (ToE), "requirement" and "check" in the UI, not "control" and "metric".

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
