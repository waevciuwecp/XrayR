# VLESS / REALITY maintenance

The compatibility workflow tests the locked module and current Xray-core main on
PRs, main pushes, weekly runs and manual runs. The upstream job logs its resolved
module version. No sibling checkout or local module replacement is used.
A passing upstream probe does not upgrade the shipped dependency (v1.260327.0).

## Panel contract

Xboard sends a top-level nullable decryption string. XrayR forwards it unchanged
to the inbound-level core configuration. Absent, null and empty values retain
none. The core validates nonempty values; this is not per-user encryption.
Encryption requires EnableFallback: false. REALITY handshake targets are separate
from VLESS fallbacks and remain supported. Panel REALITY configuration works with
DisableLocalREALITYConfig: true without a local REALITYConfigs object. IPv6 target
hosts use bracket notation when joined with their port.

The handoff's vlessenc.native examples are not a wire-format definition. The
inspected core parser requires the mlkem768x25519plus prefix. XrayR must not
rewrite or generate key strings. Panel-side enabled-with-empty-key validation
is still necessary because this API does not convey the enabled flag.

## Reference baseline

- [Panel producer](https://github.com/cedar2025/Xboard/blob/e68fec50ef48d86f7666c5ff5e758f2a6e9fa313/app/Services/ServerService.php)
- [Core parser](https://github.com/XTLS/Xray-core/blob/a4a3dff5b20c2faae567b742c172c7b95e75c083/infra/conf/vless.go)

## Upgrade gates

Prepare a pinned-core upgrade PR from upstream CI evidence and resolve API
changes before changing the release dependency. Review XHTTP options, REALITY
post-quantum signing and fallback rate limits against both panel and core schemas
before exposing additional configuration. Panel key generation and client
subscription compatibility require separate panel work.

Tests build configuration without listeners or production access. They do not
prove network handshakes, hot reload, client interoperability or performance.
Remote CI is authoritative. No deployment is included.
