# Offline-guarantee fixture

A self-contained two-file repository used by the `offline-guarantee` CI job,
which runs terrain's core commands inside a network namespace with no
interfaces, no routes, and no DNS.

The prompt deliberately references a misspelled schema field (`acount_id` vs
`account_id`) so every command under test has real work to do rather than
short-circuiting on an empty repo:

- bare `terrain` maps 1 prompt and 1 schema and surfaces the drift
- `terrain analyze` reports one high-severity `aiPromptSchemaDrift` signal
- `terrain test --fail-on high` gates on it (documented severity-gate exit 6)
- `terrain fix` proposes the one-character correction and, as a dry run,
  writes nothing

`check.sh` holds the assertions. CI copies the `.py` files into a scratch
directory and runs the script under `sudo unshare -n`; `make offline-check`
runs the same script locally with the network up (macOS has no `unshare`).
The copy step keeps the checkout clean — the commands write report state
(`.terrain/`, JUnit XML) next to the analyzed repo — and lets the script
prove the fix dry-run left the pristine source untouched.

Keep this fixture stable: the assertions depend on the exact drift it seeds.
