---
name: stop-demo
description: Stop the Confirmate demo
---

# Stop Demo

Stop the running Confirmate demo by killing the processes.

```bash
cd /Users/banse/Repositories/confirmate && bash demo.sh --stop
```

This reads the PIDs from `logs/demo.pids` and kills each process.