---
name: start-demo
description: Start the Confirmate demo in the background
---

# Start Demo

Start the Confirmate demo in daemon mode (background).

```bash
cd /Users/banse/Repositories/confirmate && bash demo.sh --daemon
```

The demo runs:
- Confirmate API on http://localhost:8080
- UI on http://localhost:5173
- code-analysis collector

Logs are written to `logs/confirmate.log`, `logs/ui.log`, and `logs/code-analysis.log`.
