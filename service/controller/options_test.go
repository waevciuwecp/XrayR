package controller_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/common/mylego"
	. "github.com/XrayR-project/XrayR/service/controller"
	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/proxy/freedom"
	"github.com/xtls/xray-core/transport/internet/reality"
	"github.com/xtls/xray-core/transport/internet/splithttp"
)

func TestRealityOptions(t *testing.T) {
	for _, panel := range []bool{false, true} {
		key := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
		seed := base64.RawURLEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
		limit := conf.LimitFallback{AfterBytes: 1024, BytesPerSec: 4096, BurstBytesPerSec: 8192}
		node := &api.NodeInfo{NodeType: "Vless", Port: 443, TransportProtocol: "tcp", EnableTLS: true, EnableREALITY: true,
			REALITYConfig: &api.REALITYConfig{Target: "127.0.0.1:8443", Dest: "legacy.invalid:443", ServerNames: []string{"example.com"}, PrivateKey: key, ShortIds: []string{"abcd"}, Mldsa65Seed: seed, LimitFallbackUpload: limit, LimitFallbackDownload: limit}}
		cfg := &Config{DisableLocalREALITYConfig: panel, EnableREALITY: true, REALITYConfigs: &REALITYConfig{Target: "127.0.0.1:8443", Dest: "legacy.invalid:443", ServerNames: []string{"example.com"}, PrivateKey: key, ShortIds: []string{"abcd"}, Mldsa65Seed: seed, LimitFallbackUpload: limit, LimitFallbackDownload: limit}}
		inbound, err := InboundBuilder(cfg, node, "options")
		if err != nil {
			t.Fatal(err)
		}
		receiver, err := inbound.ReceiverSettings.GetInstance()
		if err != nil {
			t.Fatal(err)
		}
		settings, err := receiver.(*proxyman.ReceiverConfig).StreamSettings.SecuritySettings[0].GetInstance()
		if err != nil {
			t.Fatal(err)
		}
		got := settings.(*reality.Config)
		if got.Dest != "127.0.0.1:8443" || len(got.Mldsa65Seed) != 32 || got.LimitFallbackUpload.BytesPerSec != 4096 || got.LimitFallbackDownload.AfterBytes != 1024 {
			t.Fatal("REALITY options did not reach core")
		}
		node.REALITYConfig.Mldsa65Seed = "invalid"
		cfg.REALITYConfigs.Mldsa65Seed = "invalid"
		if _, err := InboundBuilder(cfg, node, "invalid"); err == nil {
			t.Fatal("invalid seed accepted")
		}
	}
}

func TestXHTTPExtra(t *testing.T) {
	node := &api.NodeInfo{NodeType: "Vless", Port: 443, TransportProtocol: "xhttp", Path: "/upload", Host: "edge.example", XHTTPMode: "packet-up", XHTTPExtra: json.RawMessage(`{"uplinkHTTPMethod":"POST","xPaddingBytes":"100-200","scMaxBufferedPosts":17}`)}
	cfg := &Config{CertConfig: &mylego.CertConfig{CertMode: "none"}}
	inspect := func() *splithttp.Config {
		inbound, err := InboundBuilder(cfg, node, "xhttp")
		if err != nil {
			t.Fatal(err)
		}
		receiver, err := inbound.ReceiverSettings.GetInstance()
		if err != nil {
			t.Fatal(err)
		}
		settings, err := receiver.(*proxyman.ReceiverConfig).StreamSettings.TransportSettings[0].Settings.GetInstance()
		if err != nil {
			t.Fatal(err)
		}
		return settings.(*splithttp.Config)
	}
	got := inspect()
	if got.Host != "edge.example" || got.Path != "/upload" || got.Mode != "packet-up" || got.UplinkHTTPMethod != "POST" || got.ScMaxBufferedPosts != 17 || got.XPaddingBytes.From != 100 {
		t.Fatalf("XHTTP options lost: %v", got)
	}
	cfg.XHTTPMode = "stream-up"
	cfg.XHTTPExtraJSON = `{"scMaxBufferedPosts":21}`
	got = inspect()
	if got.Mode != "stream-up" || got.ScMaxBufferedPosts != 21 {
		t.Fatal("local override lost")
	}
	cfg.XHTTPExtraJSON = `{"sessionIDPlacement":"INVALID"}`
	if _, err := InboundBuilder(cfg, node, "invalid"); err == nil {
		t.Fatal("invalid XHTTP option accepted")
	}
}

func TestMissingTLSConfig(t *testing.T) {
	node := &api.NodeInfo{NodeType: "Vless", Port: 443, TransportProtocol: "tcp", EnableTLS: true}
	if _, err := InboundBuilder(&Config{}, node, "tls"); err == nil {
		t.Fatal("missing TLS configuration accepted")
	}
}

func TestFreedomFinalRules(t *testing.T) {
	ips := conf.StringList{"127.0.0.0/8"}
	cfg := &Config{FreedomFinalRules: []*conf.FreedomFinalRuleConfig{{Action: "allow", IP: &ips}}}
	outbound, err := OutboundBuilder(cfg, &api.NodeInfo{NodeType: "Vless"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	message, err := outbound.ProxySettings.GetInstance()
	if err != nil {
		t.Fatal(err)
	}
	got := message.(*freedom.Config)
	if len(got.FinalRules) != 1 || got.FinalRules[0].Action != freedom.RuleAction_Allow {
		t.Fatal("Freedom rule did not reach core")
	}
	cfg.FreedomFinalRules[0].Action = "invalid"
	if _, err := OutboundBuilder(cfg, &api.NodeInfo{NodeType: "Vless"}, "test"); err == nil {
		t.Fatal("invalid rule accepted")
	}
}
