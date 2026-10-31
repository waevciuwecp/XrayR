# modusnyan/XrayR review

Static source review on 2026-10-31 at
[86601f1](https://github.com/modusnyan/XrayR/tree/86601f10fb98509b279684027800753878c8c221).
No runtime, reload or load testing was performed. No source was copied.

## Useful maintenance ideas

- config/validate.go: aggregates errors with config paths and suggestions,
  validates panel/node combinations and REALITY short IDs, and distinguishes
  warnings from errors. This is the best candidate for a separate focused PR.
- panel/registry.go: centralizes adapter aliases and capabilities. Xboard resolves
  to the REST NewV2board adapter. This avoids divergent alias switches.
- cmd/doctor.go and observability/: preflight diagnosis and metrics are useful
  operational features; inspect their network access, authentication and exposed
  labels before adoption.
- internal/configui/: configuration editing with a redacted review screen is
  useful for operator experience, but introduces a terminal UI dependency and
  is outside the immediate protocol compatibility work.

## Limits affecting adoption

- go.mod pins Xray-core v1.251015.0, older than our v1.260327.0. This fork is not
  a source for the current core API upgrade.
- api/newV2board/model.go has no decryption field and inboundbuilder.go still
  hardcodes Decryption to none. It does not solve Xboard VLESS Encryption.
- Panel REALITY mode still dereferences config.REALITYConfigs.Show. Its separate
  CertConfig nil guard is useful, but does not address this REALITY nil pointer.
- cmd/root.go closes the old panel before starting the candidate. It attempts
  rollback after failure, but this is not uninterrupted reload. The failed
  candidate's resources need scrutiny before rollback; do not adopt based only
  on the transaction/rollback description in its commit message.
- internal/snapshot/snapshot.go removes the nested REALITY private key before
  persisting node state. Consequently that snapshot alone cannot restore panel
  REALITY credentials. It also persists user records; checksum integrity and
  restrictive file permissions do not make it encrypted credential storage.

Recommended order: protocol correctness and CI first, then isolated configuration
validation and alias cleanup, then diagnostics. Evaluate snapshots and reload as
separate reliability projects with explicit failure-path tests in remote CI.
