package newV2board

import (
	"encoding/json"
	"testing"
)

func TestVlessPanelContract(t *testing.T) {
	for _, value := range []string{`null`, `""`, `"none"`, `"mlkem768x25519plus.native.600s.test-key"`} {
		for _, host := range []string{"example.com", "2001:db8::1"} {
			t.Run(value+"/"+host, func(t *testing.T) {
				var response serverConfig
				payload := `{"server_port":443,"network":"tcp","tls":2,"flow":"xtls-rprx-vision","decryption":` + value + `,"tls_settings":{"server_name":"` + host + `","server_port":"443","private_key":"test-key","short_id":"abcd"}}`
				if err := json.Unmarshal([]byte(payload), &response); err != nil {
					t.Fatal(err)
				}
				client := &APIClient{NodeType: "Vless", EnableVless: true}
				node, err := client.parseV2rayNodeResponse(&response)
				if err != nil {
					t.Fatal(err)
				}
				var expected string
				if err := json.Unmarshal([]byte(value), &expected); err != nil {
					t.Fatal(err)
				}
				if node.Decryption != expected {
					t.Fatal("decryption was not preserved")
				}
				dest := host + ":443"
				if host == "2001:db8::1" {
					dest = "[" + host + "]:443"
				}
				if !node.EnableREALITY || node.REALITYConfig.Dest != dest || node.VlessFlow != "xtls-rprx-vision" {
					t.Fatalf("unexpected REALITY/flow mapping: %s", node.REALITYConfig.Dest)
				}
			})
		}
	}
}

func TestVlessXHTTPMapping(t *testing.T) {
	for _, network := range []string{"xhttp", "splithttp"} {
		for _, legacy := range []bool{false, true} {
			var response serverConfig
			body := `{"server_port":443,"network":"` + network + `","network_settings":{"path":"/upload","host":"edge.example","mode":"stream-up","extra":{"uplinkHTTPMethod":"POST","xPaddingBytes":"100-200"}},"tls":2,"tls_settings":{"target":"127.0.0.1:8443","mldsa65_seed":"seed","limit_fallback_upload":{"afterBytes":1024,"bytesPerSec":4096}}}`
			if err := json.Unmarshal([]byte(body), &response); err != nil {
				t.Fatal(err)
			}
			client := &APIClient{NodeType: "Vless", EnableVless: legacy}
			node, err := client.parseV2rayNodeResponse(&response)
			if err != nil {
				t.Fatal(err)
			}
			if node.Host != "edge.example" || node.Path != "/upload" || node.XHTTPMode != "stream-up" || len(node.XHTTPExtra) == 0 {
				t.Fatal("native VLESS XHTTP settings lost")
			}
			if node.REALITYConfig.Target != "127.0.0.1:8443" || node.REALITYConfig.Mldsa65Seed != "seed" || node.REALITYConfig.LimitFallbackUpload.BytesPerSec != 4096 {
				t.Fatal("REALITY options lost")
			}
		}
	}
}

func TestVlessLegacyNetworkSettings(t *testing.T) {
	var response serverConfig
	if err := json.Unmarshal([]byte(`{"network":"xhttp","networkSettings":{"path":"/legacy","host":"legacy.example"}}`), &response); err != nil {
		t.Fatal(err)
	}
	node, err := (&APIClient{NodeType: "Vless"}).parseV2rayNodeResponse(&response)
	if err != nil {
		t.Fatal(err)
	}
	if node.Path != "/legacy" || node.Host != "legacy.example" {
		t.Fatal("legacy networkSettings lost")
	}
}
