# Xboard Hysteria 2 with XrayR

XrayR accepts Xboard Hysteria nodes with `version: 2`. `Hysteria2` is also accepted as a node type alias and is sent to XrayR's Hysteria controller. Hysteria 1 is not supported.

Before enabling the node:

- Configure `ControllerConfig.CertConfig` with a certificate and key. `CertMode: none` is rejected because Hysteria 2 requires TLS. The node's panel `cert_config` is not used by XrayR.
- XrayR configures the required `h3` ALPN for Hysteria 2, independently of the controller's general TLS ALPN setting.
- XrayR maps Xboard Hysteria 2 `obfs: salamander` and `obfs-password` to Xray-core's UDP FinalMask. The password must be at least 4 bytes. When `obfs` is empty, XrayR leaves obfuscation disabled even if Xboard retains an `obfs-password` value. Other obfuscation types, including Gecko, are rejected with a node-specific error.
- Controller `FinalMaskJSON` is preserved; the node's Salamander mask is added before any UDP `filter` mask so the filter remains outermost.
- Treat panel `up_mbps` and `down_mbps` as informational for this backend. XrayR does not apply these values as Hysteria 2 bandwidth limits.

XrayR uses each panel user's UUID as the Hysteria 2 authentication password, matching Xboard's subscription output.
