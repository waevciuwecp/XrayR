# Xboard Hysteria 2 with XrayR

XrayR accepts Xboard Hysteria nodes with `version: 2`. `Hysteria2` is also accepted as a node type alias and is sent to XrayR's Hysteria controller. Hysteria 1 is not supported.

Before enabling the node:

- Configure `ControllerConfig.CertConfig` with a certificate and key. `CertMode: none` is rejected because Hysteria 2 requires TLS. The node's panel `cert_config` is not used by XrayR.
- XrayR configures the required `h3` ALPN for Hysteria 2, independently of the controller's general TLS ALPN setting.
- Leave panel obfuscation disabled. Xray-core does not implement Hysteria 2 obfuscation, so XrayR rejects a node configuration that enables it.
- Treat panel `up_mbps` and `down_mbps` as informational for this backend. XrayR does not apply these values as Hysteria 2 bandwidth limits.

XrayR uses each panel user's UUID as the Hysteria 2 authentication password, matching Xboard's subscription output.
