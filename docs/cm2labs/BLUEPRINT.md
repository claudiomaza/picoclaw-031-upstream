# BLUEPRINT: PicoClaw extensibility fixes

## Goals
- Add a read-only `vm1_inspect` tool with fixed allowlisted targets and runtime-root discovery.
- Add channel-independent rendering profiles, with WhatsApp normalization at the WhatsApp adapter boundary.
- Preserve upstream behavior through cm2labs extension seams.

## Safety
- No arbitrary paths, writes, target-binary execution, shell commands, or secret reads.
- Reject traversal and symlink escapes.
- Inspection output is structured and redacts sensitive filenames/content.

## Scope
- `pkg/extensions/formatting/`: canonical Markdown to channel profiles.
- `pkg/extensions/inspector/`: discovery, allowlist, metadata, hashes, safe strings, systemd status.
- `pkg/tools/inspector/`: agent-facing read-only adapter.
- `pkg/channels/whatsapp/`: invoke WhatsApp profile only.
- Tests for formatting, paths, symlinks, metadata and tool contract.

## Verification
- gofmt, go test ./..., go vet ./..., build, diff check, runtime smoke test.
