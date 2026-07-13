// Custom Node server that serves the SvelteKit adapter-node build
// and proxies /v1/ requests to the core API.
import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { extname, join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const PORT = process.env.PORT || 80;
const API_BASE = process.env.API_BASE || 'http://core:8080';
const BUILD_DIR = join(__dirname, 'build');

const MIME = {
  '.html': 'text/html',
  '.js': 'application/javascript',
  '.mjs': 'application/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.map': 'application/json',
  '.txt': 'text/plain',
};

const server = createServer(async (req, res) => {
  const url = new URL(req.url, `http://localhost:${PORT}`);
  const pathname = url.pathname;

  // Proxy API requests to core
  if (pathname.startsWith('/v1/')) {
    try {
      const headers = { ...req.headers };
      delete headers.host;
      
      const proxyRes = await fetch(`${API_BASE}${pathname}${url.search}`, {
        method: req.method,
        headers,
        body: ['GET', 'HEAD'].includes(req.method) ? undefined : req,
        redirect: 'manual',
      });

      const respHeaders = {};
      proxyRes.headers.forEach((v, k) => { respHeaders[k] = v; });
      respHeaders['Access-Control-Allow-Origin'] = '*';
      
      res.writeHead(proxyRes.status, respHeaders);

      // Stream the response body
      const buf = Buffer.from(await proxyRes.arrayBuffer());
      res.end(buf);
    } catch (e) {
      console.error('Proxy error:', e.message);
      res.writeHead(502, { 'Content-Type': 'text/plain' });
      res.end('Bad Gateway');
    }
    return;
  }

  // Serve static assets from build/client
  if (pathname.startsWith('/_app/')) {
    const filePath = join(BUILD_DIR, 'client', pathname);
    try {
      const data = await readFile(filePath);
      const ext = extname(filePath);
      res.writeHead(200, {
        'Content-Type': MIME[ext] || 'application/octet-stream',
        'Cache-Control': 'public, max-age=31536000, immutable',
      });
      res.end(data);
      return;
    } catch {
      // fall through to 404
    }
  }

  // For all other routes, serve the SPA (client-side routing)
  // The adapter-node build generates the HTML dynamically via the server handler.
  // Since we can't run the full adapter-node server (it needs Node 22),
  // we serve a minimal HTML shell that loads the client-side app.
  try {
    // Try to find index.html in various locations
    const possiblePaths = [
      join(BUILD_DIR, 'client', 'index.html'),
      join(BUILD_DIR, 'index.html'),
    ];
    
    for (const p of possiblePaths) {
      try {
        const html = await readFile(p, 'utf-8');
        res.writeHead(200, { 'Content-Type': 'text/html' });
        res.end(html);
        return;
      } catch {
        // try next
      }
    }
    
    // No index.html found — generate a minimal SPA shell
    const appJs = await findAppEntry();
    const html = generateSpaShell(appJs);
    res.writeHead(200, { 'Content-Type': 'text/html' });
    res.end(html);
  } catch (e) {
    console.error('Serve error:', e.message);
    res.writeHead(500, { 'Content-Type': 'text/plain' });
    res.end('Internal Server Error');
  }
});

async function findAppEntry() {
  // Look for the main JS bundle in _app/
  try {
    const dir = join(BUILD_DIR, 'client', '_app');
    const { readdir } = await import('node:fs/promises');
    
    // Check version.json for the app version
    try {
      const versionData = JSON.parse(await readFile(join(dir, 'version.json'), 'utf-8'));
      const version = versionData.version;
      if (version) {
        return `/_app/version.json`;
      }
    } catch {}
    
    return null;
  } catch {
    return null;
  }
}

function generateSpaShell(appEntry) {
  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Confirmate</title>
  <link rel="stylesheet" href="/_app/immutable/assets/app.css" />
</head>
<body>
  <div id="svelte"></div>
  <script type="module" src="/_app/immutable/entry/start.js"></script>
</body>
</html>`;
}

server.listen(PORT, () => {
  console.log(`UI server running on port ${PORT}, proxying /v1/ to ${API_BASE}`);
});
