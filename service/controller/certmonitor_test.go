package controller

import (
	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/common/mylego"
	"testing"
)

func TestCertificateMonitorReality(t *testing.T) {
	for _, tc := range []struct {
		name   string
		node   *api.NodeInfo
		config *Config
		renew  bool
	}{
		{name: "no-cert", node: &api.NodeInfo{EnableTLS: true}, config: &Config{}},
		{name: "panel-reality", node: &api.NodeInfo{EnableTLS: true, EnableREALITY: true, REALITYConfig: &api.REALITYConfig{}}, config: &Config{DisableLocalREALITYConfig: true, CertConfig: &mylego.CertConfig{CertMode: "dns"}}},
		{name: "local-reality", node: &api.NodeInfo{EnableTLS: true}, config: &Config{EnableREALITY: true, REALITYConfigs: &REALITYConfig{}, CertConfig: &mylego.CertConfig{CertMode: "dns"}}},
		{name: "plain", node: &api.NodeInfo{}, config: &Config{CertConfig: &mylego.CertConfig{CertMode: "dns"}}},
		{name: "static", node: &api.NodeInfo{EnableTLS: true}, config: &Config{CertConfig: &mylego.CertConfig{CertMode: "file"}}},
		{name: "tls-dns", node: &api.NodeInfo{EnableTLS: true}, config: &Config{CertConfig: &mylego.CertConfig{CertMode: "dns"}}, renew: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Controller{nodeInfo: tc.node, config: tc.config}
			if c.needsCertificateRenewal() != tc.renew {
				t.Fatal("wrong certificate renewal decision")
			}
			if !tc.renew {
				if err := c.certMonitor(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
