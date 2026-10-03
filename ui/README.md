# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```sh
# create a new project
npx sv create my-app
```

To recreate this project with the same configuration:

```sh
# recreate this project
npx sv@0.13.0 create --template minimal --types ts --install npm ui
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```sh
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```sh
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

## Serving under a path prefix

To serve the UI below a path prefix (e.g. behind a reverse proxy at
`https://<host>/proxy/5173/`), build it with `UI_BASE_PATH`:

```sh
UI_BASE_PATH=/proxy/5173 npm run build
```

The value must start with `/` and must not end with one. Internal links must
use `resolve()` from `$app/paths` so they carry the prefix. For the embedded
login, see "Running behind a path-prefix reverse proxy" in
`core/docs/authentication-and-authorization.md`.

