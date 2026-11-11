package newV2board

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/XrayR-project/XrayR/api"
)

func TestHysteria2PanelContract(t *testing.T) {
	var configRequests, userRequests int
	var requestedNodeTypes []string
	panel := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedNodeTypes = append(requestedNodeTypes, r.URL.Query().Get("node_type"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/server/UniProxy/config":
			configRequests++
			_, _ = w.Write([]byte(`{"server_port":443,"version":2,"host":"hy2.example.com","server_name":"hy2.example.com","network":null,"obfs":null}`))
		case "/api/v1/server/UniProxy/user":
			userRequests++
			_, _ = w.Write([]byte(`{"users":[{"id":9,"uuid":"user-uuid"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer panel.Close()

	for _, nodeType := range []string{"Hysteria", "Hysteria2"} {
		client := New(&api.Config{APIHost: panel.URL, NodeID: 14, NodeType: nodeType})
		node, err := client.GetNodeInfo()
		if err != nil {
			t.Fatalf("GetNodeInfo(%s): %v", nodeType, err)
		}
		if node.NodeType != "Hysteria" || node.Port != 443 || node.TransportProtocol != "hysteria" || !node.EnableTLS {
			t.Fatalf("unexpected Hysteria 2 node mapping: %+v", node)
		}
		if node.Host != "hy2.example.com" || node.ServiceName != "hy2.example.com" {
			t.Fatalf("Hysteria 2 host/SNI fields lost: %+v", node)
		}

		users, err := client.GetUserList()
		if err != nil {
			t.Fatalf("GetUserList(%s): %v", nodeType, err)
		}
		if len(*users) != 1 || (*users)[0].UUID != "user-uuid" {
			t.Fatalf("unexpected Hysteria 2 users: %+v", *users)
		}
	}

	if configRequests != 2 || userRequests != 2 {
		t.Fatalf("unexpected panel requests: config=%d users=%d", configRequests, userRequests)
	}
	for i, want := range []string{"hysteria", "hysteria", "hysteria2", "hysteria2"} {
		if requestedNodeTypes[i] != want {
			t.Fatalf("request %d used node_type %q, want %q", i, requestedNodeTypes[i], want)
		}
	}
}

func TestHysteria2ObfsPanelSettings(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantErr      string
		wantObfsType string
		wantPassword string
	}{
		{name: "v1", body: `{"server_port":443,"version":1}`, wantErr: "only Hysteria 2"},
		{name: "salamander", body: `{"server_port":443,"version":2,"obfs":"salamander","obfs-password":"long-enough-password"}`, wantObfsType: "salamander", wantPassword: "long-enough-password"},
		{name: "obfs disabled with retained password", body: `{"server_port":443,"version":2,"obfs":null,"obfs-password":"stale-password"}`},
		{name: "short salamander password", body: `{"server_port":443,"version":2,"obfs":"salamander","obfs-password":"abc"}`, wantErr: "at least 4 bytes"},
		{name: "unsupported obfs", body: `{"server_port":443,"version":2,"obfs":"gecko"}`, wantErr: `unsupported Hysteria 2 obfs "gecko"`},
		{name: "invalid port", body: `{"server_port":65536,"version":2}`, wantErr: "invalid server_port"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response serverConfig
			if err := json.Unmarshal([]byte(tt.body), &response); err != nil {
				t.Fatal(err)
			}
			node, err := (&APIClient{NodeID: 1}).parseHysteriaNodeResponse(&response)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("got error %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseHysteriaNodeResponse: %v", err)
			}
			if node.ObfsType != tt.wantObfsType || node.ObfsPassword != tt.wantPassword {
				t.Fatalf("got obfs type/password %q/%q, want %q/%q", node.ObfsType, node.ObfsPassword, tt.wantObfsType, tt.wantPassword)
			}
		})
	}
}
