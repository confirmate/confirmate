# UI design guide

This guide explains how the user interface of Confirmate is designed and built. It is for contributors, human or
AI. The short rules that must always be followed are in [AGENTS.md](../../AGENTS.md); this document gives the reasons
and the details.

The guided experience (`ui/src/routes/guided/`) follows this guide. The older screens under `routes/(global)` and
`routes/(toe)` do not yet; they are replaced step by step.

## Principles

1. **Task first.** A screen answers "what do I have to do next?", not "which backend entity is this?". Entity-centric
   views (evidence, assessment results) exist for experts and are kept out of the main path.
2. **Plain words.** Use the vocabulary of the user, not of the data model (table below). The expert area keeps the
   technical names.
3. **Guide, then get out of the way.** New users get a wizard with a sensible default at every step. Experienced
   users get direct access to every list and detail.
4. **Never silent.** Anything that reduces the scope or changes a compliance statement needs a reason and a person
   who approves it. Unknown input never reduces the scope.
5. **Status is never only a color.** Every status has a text label and, where useful, an icon. Colors that must be
   told apart differ in lightness, not only in hue.
6. **Accessible by default.** Real `<button>`, `<a>`, `<label>` and `<input>` elements, visible focus, targets of at
   least 44 px in primary flows, text contrast of at least 4.5:1, layouts that work at phone width.

## Terminology

| Shown to the user | In the backend |
|-------------------|----------------|
| Target of Evaluation (ToE) with a kind (product, cloud service, organisation) | target of evaluation |
| Requirement | control |
| Check | metric |
| Check result | assessment result |
| Requirement status | evaluation result |
| Framework (CRA, BSI C5, NIS-2) | catalog |
| Compliance for a framework | audit scope |

"Target of Evaluation" is kept on purpose, because it is established in the compliance domain. A framework profile
may add its own word where it reads better, for example "product" on the CRA pages.

## Structure of the code

```
ui/src/lib/theme/      design tokens and themes (theme.css), theme state
ui/src/lib/ds/         design system: generic building blocks without domain knowledge
ui/src/lib/guided/     the guided experience
  profile.ts           what a framework profile is
  profiles/            the profiles (plain data), one per framework
  source.ts, mock.ts   where the screens get their data from, and the sample data
  components/          page level and domain components
ui/src/routes/guided/  route files
```

- **Route files are the bare minimum.** A `+page.svelte` or `+layout.svelte` passes the loaded data to one component
  and contains no markup of its own. Loading and redirects stay in `+page.ts`/`+layout.ts`. If a route file grows
  beyond a few lines, move the markup into a component in `$lib/guided/components`.
- **Reuse before you write.** Look in `$lib/ds` and `$lib/guided/components` first. If something is used twice, make
  it a component; if it is generic and knows nothing about compliance, it belongs in `$lib/ds`.
- **No domain knowledge in `$lib/ds`.** It must not import from `$lib/guided` or from the API types.
- **No framework specifics in generic components.** What differs between the CRA, C5 and NIS-2 is data in a profile
  (steps, questions, scoping effects, sources, vocabulary), not branches in a component.
- **Data only through `GuidedSource`.** Components do not call the API. A mock implements `GuidedSource` until the
  backend pieces exist; the real implementation replaces it without changing a screen.

## Design tokens and themes

All colors, fonts and radii of the guided experience come from tokens in `ui/src/lib/theme/theme.css`.

- A **theme** is a block of `--cf-*` variable values selected by `data-theme` on `<html>`. Add a theme by adding a
  block with the same variables and listing it in `theme.svelte.ts`.
- Components use only the **semantic utilities** that the tokens generate: `bg-canvas`, `bg-surface`, `bg-sunken`,
  `text-ink`, `text-ink-soft`, `text-ink-muted`, `border-line`, `bg-brand`, `text-on-brand`, `bg-brand-soft`,
  `text-ok`/`bg-ok-soft`/`text-ok-ink` (and the same for `warn` and `bad`), `rounded-control`, `rounded-card`,
  `font-body`, `font-display`, `font-code`. Do not use raw palette classes such as `bg-gray-100` or `text-blue-600`
  in new components, and do not write colors in hex.
- Tones (neutral, brand, ok, warn, bad) are expressed through `tones.ts`, so that a status looks the same everywhere.
- Fonts are self-hosted through npm packages (the product runs self-hosted and offline); do not load fonts from a CDN.
- Spacing and size use the Tailwind scale. Keep line length of body text below about 70 characters.

## Building blocks

| Component | Use it for |
|-----------|------------|
| `Button` | Actions. One primary action per screen area, others `quiet`. Links that look like buttons use `href`. |
| `Chip` | Short labels and statuses (with a text, never a color alone). |
| `Card` | A group of related content. `accent` marks status with a left border. |
| `Field`, `TextInput`, `Select` | Form fields with label and hint. |
| `OptionCard` | One choice out of a few, with an explanation. |
| `Segmented` | A short choice such as yes, no, don't know. |
| `FilterChips` | Filtering a list by one attribute. |
| `Stepper` | The steps of a wizard. |
| `StackedProgress` | Progress with several outcomes. Always next to the numbers. |
| `Banner` | Something that needs attention or explains a consequence. |
| `PageHeading` | Title, introduction and actions of a page. |

## Patterns

- **Wizard.** Steps come from a profile and use the fixed step types (details, choice, questionnaire, scope,
  sources). Every step has a title, one sentence of introduction, a default, and a way back. A side panel shows the
  consequence of the answers as they change (for example how many requirements apply). Progress is kept when the
  person leaves.
- **Work queue.** A list of tasks with a status, an explanation of why, who is responsible and when. Each task links
  to the place where it is solved.
- **Requirement view.** The requirement in plain words and its reference, the status, the checks with the evidence
  behind them, and what a person can do. Automatic results are shown as such and win over statements.
- **Scope decisions.** Always show the reason, who decided and when; require a confirmation.
- **Empty and error states.** Say what is missing and what to do about it.

## Adding a framework profile

1. Create `ui/src/lib/guided/profiles/<name>.ts` with a `FrameworkProfile` and add it to `profiles/index.ts`.
2. Use the existing step types. If a framework needs a new kind of step, add a generic step component and a step
   type, not a special case.
3. Keep every text of the framework in the profile. Texts that are not reviewed yet are marked as sample text.
4. Check the result in the browser with the feature flag on (see below).

## Trying it

The guided experience is off by default. Switch it on with `VITE_FEATURE_GUIDED=true` at build time, or in the
browser with `localStorage.setItem('cf.feature.guided', 'true')`, then open `/guided/`.

```
cd ui && VITE_FEATURE_GUIDED=true npm run dev     # with an orchestrator on port 8080
```

Before a pull request: `npm run check` shows no errors in the files you touched and `npm run build` succeeds.
