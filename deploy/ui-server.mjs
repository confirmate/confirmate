import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { extname, join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const PORT = process.env.PORT || 80;
const API_BASE = process.env.API_BASE || 'http://core:8080';
const STATIC_DIR = process.env.STATIC_DIR || join(__dirname, 'static');

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
      delete headers['content-length'];

      // Buffer the request body for POST/PUT/PATCH
      let body = undefined;
      if (!['GET', 'HEAD'].includes(req.method)) {
        const chunks = [];
        for await (const chunk of req) {
          chunks.push(chunk);
        }
        body = Buffer.concat(chunks);
      }
      
      const proxyRes = await fetch(`${API_BASE}${pathname}${url.search}`, {
        method: req.method,
        headers,
        body,
        redirect: 'manual',
      });

      const respHeaders = {};
      proxyRes.headers.forEach((v, k) => { 
        // Fix relative Location headers to point through the proxy
        if (k.toLowerCase() === 'location' && v.startsWith('/')) {
          respHeaders[k] = v;
        } else {
          respHeaders[k] = v;
        }
      });
      respHeaders['Access-Control-Allow-Origin'] = '*';
      
      res.writeHead(proxyRes.status, respHeaders);
      const buf = Buffer.from(await proxyRes.arrayBuffer());
      res.end(buf);
    } catch (e) {
      console.error('Proxy error:', e.message);
      res.writeHead(502, { 'Content-Type': 'text/plain' });
      res.end('Bad Gateway');
    }
    return;
  }

  // Serve static files
  let filePath = join(STATIC_DIR, pathname);
  if (pathname === '/' || !extname(pathname)) {
    filePath = join(STATIC_DIR, 'index.html');
  }

  try {
    const data = await readFile(filePath);
    const ext = extname(filePath);
    res.writeHead(200, { 'Content-Type': MIME[ext] || 'application/octet-stream' });
    res.end(data);
  } catch {
    // SPA fallback
    try {
      const index = await readFile(join(STATIC_DIR, 'index.html'));
      res.writeHead(200, { 'Content-Type': 'text/html' });
      res.end(index);
    } catch {
      res.writeHead(404, { 'Content-Type': 'text/plain' });
      res.end('Not Found');
    }
  }
});

server.listen(PORT, () => {
  console.log(`UI server running on port ${PORT}, proxying /v1/ to ${API_BASE}`);
});
