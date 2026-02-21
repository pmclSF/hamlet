# ADR-004: Jest ESM Strategy and Worker Shutdown Warning

## Status

Accepted (defer migration)

## Context

Hamlet is an ES modules project (`"type": "module"`). Jest 29 does not natively
support ESM; it requires `NODE_OPTIONS='--experimental-vm-modules'` to load
`.js` test files as ES modules via Node's VM module API.

This works correctly for test execution but has a side-effect: the experimental
VM modules loader creates internal IPC Socket handles inside each Jest worker
process. These handles are not visible to `--detectOpenHandles` (they exist at
the C++/libuv level, not as JavaScript-level refs) and prevent the worker's
event loop from draining within Jest's shutdown grace period (~5 s).

When the last worker to finish does not exit in time, Jest force-kills it and
prints:

> A worker process has failed to exit gracefully and has been force exited.

This warning is **systemic** — it is not caused by any individual test file. It
was confirmed by excluding the slowest test (`cli.test.js`) and observing the
identical warning for the next-slowest worker.

## Options Considered

### 1. Upgrade to Jest 30

Jest 30 ships improved ESM support and may resolve the worker handle issue.
However, Jest 30 is a major version bump with potential breaking changes across
706 test suites. This requires dedicated migration effort and testing.

### 2. Migrate to Vitest

Vitest has native ESM support with no experimental flags. Migration would
eliminate the warning entirely but requires rewriting test infrastructure
(configuration, coverage setup, CI pipelines) and potentially adjusting test
syntax.

### 3. Custom test environment with `handle.unref()`

A custom `TestEnvironment` subclass that calls `handle.unref()` on all active
handles during `teardown()` was tested. It caused workers to exit before Jest
could collect test results, crashing 268 of 706 suites. This approach is not
viable without deeper integration into Jest's worker lifecycle.

### 4. `forceExit: true` in jest.config.js

This config option tells Jest's main process to call `process.exit()` after all
results are collected. However, it does not prevent the *worker-level* warning —
the worker is already force-killed before the main process exits. Both warnings
appear simultaneously.

### 5. Defer (document and accept)

Accept the warning as a known Jest 29 limitation. Document it in `jest.config.js`
and `CLAUDE.md` so contributors are not alarmed. Revisit when Jest 30 is
evaluated for adoption.

## Decision

**Option 5 — Defer.** The warning is cosmetic and does not affect test
correctness, coverage collection, or CI pass/fail status. All 706 suites and
1918 tests pass reliably.

Re-evaluate when:
- Jest 30 is assessed for compatibility with the project
- The warning begins causing CI false-negatives (e.g., a CI step that greps for
  "fail" in output)
- A new Jest release resolves the `--experimental-vm-modules` handle leak

## References

- [Jest ESM docs](https://jestjs.io/docs/ecmascript-modules)
- [Node.js `--experimental-vm-modules`](https://nodejs.org/api/vm.html#vm_class_vm_module)
- Jest worker source: `node_modules/jest-worker/build/workers/ChildProcessWorker.js`
