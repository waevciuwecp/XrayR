# VLESS / REALITY follow-up

This change continues the core upgrade pinned in v26.10.31.1. It does not change
the core revision or turn protocol validation off.

## Configuration migration and startup errors

VLESS Encryption and VLESS fallback cannot be enabled together. For each encrypted
node set ControllerConfig.EnableFallback to false, or disable Encryption in the
panel. Clients must use the matching client encryption string and support that
feature. Restarting with a mismatched subscription cannot fix a handshake.

Node setup errors now propagate as ordinary errors and failed panel startup closes
controllers and the core. Startup still fails on invalid node configuration; it
neither silently skips nodes nor changes their encryption or fallback settings.
Certificate renewal uses the same local/panel REALITY selection as inbound
building and skips REALITY, static certificates and missing certificate configs.
Other legacy configuration loaders may still panic on unrelated invalid inputs.

## Mapping and new options

Native NodeType Vless no longer requires EnableVless merely to read network_settings.
Both xhttp and splithttp read Host, Path, mode and extra. Legacy networkSettings
responses remain supported when network_settings is absent/null.

XHTTP advanced options use the core's standard extra JSON object. For example:

```yaml
ControllerConfig:
  XHTTPMode: packet-up
  XHTTPExtraJSON: '{"uplinkHTTPMethod":"POST","xPaddingBytes":"100-200","scMaxBufferedPosts":16}'
  SniffingIPsExcluded:
    - 192.0.2.0/24
```

A nonempty local XHTTPMode overrides the panel mode. Nonempty XHTTPExtraJSON
replaces the panel network_settings.extra object. The core's extra semantics apply:
Host/Path/Mode remain outside extra; other options inside extra replace the legacy
individual XHTTP knobs. Invalid JSON or core-invalid values fail config building.
Only configure options relevant to the endpoint; client downloadSettings/xmux
configuration belongs in subscriptions, not a server node.

Local REALITY configuration additionally accepts:

```yaml
ControllerConfig:
  EnableREALITY: true
  DisableLocalREALITYConfig: false
  REALITYConfigs:
    Target: example.com:443
    ServerNames: [example.com]
    PrivateKey: '<X25519 private key>'
    ShortIds: [abcd]
    Mldsa65Seed: '<separate ML-DSA-65 seed>'
    LimitFallbackUpload:
      AfterBytes: 1048576
      BytesPerSec: 65536
      BurstBytesPerSec: 131072
    LimitFallbackDownload:
      AfterBytes: 1048576
      BytesPerSec: 65536
      BurstBytesPerSec: 131072
```

These are illustrative rates, not deployment recommendations. Target takes
precedence over legacy Dest. ML-DSA-65 is optional, needs a paired client verify
key, and needs enough space in the target certificate flight for its signature.
REALITY fallback rate limits are distinct from VLESS EnableFallback.

Panel mode (DisableLocalREALITYConfig: true) recognizes tls_settings.target,
mldsa65_seed, limit_fallback_upload and limit_fallback_download. Rate object keys
use the core names afterBytes/bytesPerSec/burstBytesPerSec. The inspected Xboard
schema does not yet expose these fields: use local configuration or implement
panel schema, validation and subscription support before relying on panel mode
for these new capabilities. No Xboard changes are included here.

FreedomFinalRules maps to the core's freedom finalRules configuration. An omitted
SendIP remains omitted instead of passing an invalid empty sendThrough. Default
core blocks remain unchanged; loopback integration tests explicitly allow only
their local target. Do not add broad production allow rules just to emulate tests.

## Accounting and dispatch

The custom dispatcher now registers as the routing dispatcher actually consumed
by core inbounds; previously the core used a separate official dispatcher and
bypassed XrayR limits/rules. Existing configured device/speed limits and detection
rules now take effect. Both Dispatch and DispatchLink use core accounting and
online IP tracking, with context-aware shared rate limits. Limits are installed
before users are authenticated. Splice remains disabled to preserve this path.
No claim is made about throughput changes without load testing.

After a successful report, only the captured byte counts are subtracted from
atomic counters. Bytes added during the request remain for the next report;
a failed report retains all bytes. This does not make panel requests idempotent:
a response lost after the panel commits a report can still cause duplicate billing.

The reader/sniffer follows the current core's timeout and fragmented-input logic,
including composite FakeDNS route-only handling and cancellation propagation.

## Verification

Targeted tests include config propagation and rejection, concurrent report
acknowledgement, cancellation/online cleanup, second-device rejection, fragmented
headers with byte preservation, and failed-start cleanup. Real core client/server
loopback tests cover TCP, Encryption, Vision+REALITY, Vision+REALITY+ML-DSA-65, Vision+REALITY+Encryption,
XHTTP, and XHTTP+REALITY+Encryption; both direction counters must match payload
length exactly. Everything uses ephemeral local keys and loopback services.

REALITY performs asynchronous target probes. Tests wait for the actual probe
result before connecting, rather than racing its five-second readiness poll.
This is not evidence that production REALITY targets are reachable or suitable.

No production deployment, reload test or load test is included. Remote CI tests
both pinned and upstream core revisions, with a focused Linux race check on the
pinned core. PR builds remain Linux amd64; release matrix settings are unchanged.
