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

(Any further agent-specific rules can be added here.)
