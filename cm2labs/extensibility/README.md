# PicoClaw cm2labs extensibility

This directory owns cm2labs contracts, manifests and documentation.

## Compile-time Go extensions

Executable Go implementation belongs under:

```text
pkg/extensibility/
```

Code placed there must be imported by a reachable command package (`cmd/picoclaw` or `cmd/picoclaw-a2a`) to be included in the final static binary. The `cm2labs/extensibility/` directory is not loaded dynamically and is not itself a Go plugin directory.

Current integration seams:

- `pkg/extensibility/a2a/`
- `pkg/extensibility/profile/`
- `pkg/extensibility/metadata/`
- `pkg/tools/`
- `pkg/agent/`

Build examples:

```text
go build ./cmd/picoclaw
go build ./cmd/picoclaw-a2a
```

Both commands produce a single binary with the reachable cm2labs Go code statically linked.
