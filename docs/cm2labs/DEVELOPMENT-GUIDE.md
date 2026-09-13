# cm2labs runtime development contract

PicoClaw remains an independent runtime. When it is developed or changed through `agent-harness`, the harness governs the coding lifecycle, evidence and gates; it does not replace PicoClaw's runtime behavior.

## Required lifecycle

```text
Discovery → plan/documentation → isolated worktree → TDD
→ minimal implementation → verification → PR by boundary → deploy with backup/rollback
```

Before editing, identify the runtime root, host, branch, upstream/local boundary, entrypoint, configuration, persistence, service, channel and rollback artifact. Do not infer an architecture from a partial test or from an agent workspace that is not the runtime host.

## Harness task contract

A coding task must state:

- goal and non-goals;
- repository root and runtime host;
- files/boundary allowed to change;
- upstream files that must remain untouched;
- tests and acceptance evidence;
- build/deploy requirements;
- backup and rollback procedure;
- run, trace, session IDs and provenance.

`session-gateway` owns Telegram sessions. A2A is `Zapia → agent-harness → a2a-go → PicoClaw`; PicoClaw turns use `/v1/profile/agents/{agent_id}/turn` with `agent_id=default`, not `/v1/a2a/turn`.

## Change policy

Prefer adapters, hooks, seams, overlays and configuration. If an upstream edit is unavoidable, preserve visible upstream behavior, document the reason, add a reapplication test and record rollback. Never add secrets to the repository or silently change providers, models, persistence or channel ownership.

## Verification gates

At minimum, run the relevant tests, `go vet ./...`, a reproducible build and `git diff --check`. For Native WhatsApp on VM1 use the documented low-memory build flags, verify architecture/module metadata/`whatsmeow`/SHA-256, preserve the production binary and smoke-test the service before declaring readiness. Separate `implemented`, `compiled`, `configured`, `enabled` and `operational` in the report.

## Evidence returned by a completed task

Report changed files, tests and exact results, build metadata, artifact path and checksum, service state, smoke-test result, known limitations and rollback path. Do not claim full readiness from partial evidence.
