package controller_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/common/mylego"
	. "github.com/XrayR-project/XrayR/service/controller"
	vlessin "github.com/xtls/xray-core/proxy/vless/inbound"
)

func TestVlessInboundContract(t *testing.T) {
	key := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	encrypted := "mlkem768x25519plus.native.600s." + key
	for _, tc := range []struct {
		name, decryption  string
		fallback, invalid bool
	}{
		{"legacy", "", false, false},
		{"none", "none", false, false},
		{"encrypted", encrypted, false, false},
		{"encrypted-fallback", encrypted, true, true},
		{"invalid", "invalid", false, true},
		{"legacy-fallback", "", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node := &api.NodeInfo{NodeType: "Vless", NodeID: 42, Port: 443, TransportProtocol: "tcp", Decryption: tc.decryption}
			cfg := &Config{EnableFallback: tc.fallback, CertConfig: &mylego.CertConfig{CertMode: "none"}}
			result, err := InboundBuilder(cfg, node, "contract")
			if tc.invalid {
				if err == nil {
					t.Fatal("expected rejection")
				}
				if tc.fallback && !strings.Contains(err.Error(), "EnableFallback: false") {
					t.Fatal(err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			message, err := result.ProxySettings.GetInstance()
			if err != nil {
				t.Fatal(err)
			}
			inbound := message.(*vlessin.Config)
			if tc.decryption == encrypted {
				if inbound.Decryption != key || inbound.SecondsFrom != 600 {
					t.Fatal("encryption did not reach core")
				}
			} else if inbound.Decryption != "none" {
				t.Fatal("legacy default changed")
			}
		})
	}
}

func TestVlessPanelRealityWithoutLocalConfig(t *testing.T) {
	node := &api.NodeInfo{NodeType: "Vless", Port: 443, TransportProtocol: "tcp", EnableTLS: true, EnableREALITY: true,
		REALITYConfig: &api.REALITYConfig{Dest: "example.com:443", ServerNames: []string{"example.com"},
			PrivateKey: base64.RawURLEncoding.EncodeToString(make([]byte, 32)), ShortIds: []string{"abcd"}}}
	cfg := &Config{DisableLocalREALITYConfig: true, CertConfig: &mylego.CertConfig{CertMode: "none"}}
	if _, err := InboundBuilder(cfg, node, "reality"); err != nil {
		t.Fatal(err)
	}
}
