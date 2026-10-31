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
