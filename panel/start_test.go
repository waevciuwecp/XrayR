package panel

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/service/controller"
)

func TestPanelStartupRollback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"server_port":443,"network":"tcp","tls":0,"decryption":"mlkem768x25519plus.native.600s.synthetic"}`))
	}))
	defer server.Close()
	p := New(&Config{NodesConfig: []*NodesConfig{{PanelType: "NewV2board", ApiConfig: &api.Config{APIHost: server.URL, NodeID: 10, NodeType: "V2ray", EnableVless: true}, ControllerConfig: &controller.Config{EnableFallback: true}}}})
	err := p.Start()
	if err == nil || !strings.Contains(err.Error(), "EnableFallback: false") {
		t.Fatalf("expected configuration error, got %v", err)
	}
	if p.Running || p.Server != nil || len(p.Service) != 0 {
		t.Fatal("failed startup retained resources")
	}
	p.Close()
}
