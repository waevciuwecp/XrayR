package panel

import (
	"errors"
	"fmt"
	"net"
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

func TestPanelStartupKeepsHealthyNodesWhenOneNodeFails(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	goodPort := probe.Addr().(*net.TCPAddr).Port
	if err := probe.Close(); err != nil {
		t.Fatal(err)
	}

	panelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/server/UniProxy/config":
			if r.URL.Query().Get("node_id") == "10" {
				_, _ = w.Write([]byte(`{"server_port":443,"network":"tcp","tls":0,"decryption":"mlkem768x25519plus.native.600s.synthetic"}`))
				return
			}
			_, _ = fmt.Fprintf(w, `{"server_port":%d,"network":"tcp","tls":0,"decryption":"none"}`, goodPort)
		case "/api/v1/server/UniProxy/user":
			_, _ = w.Write([]byte(`{"users":[{"id":1,"uuid":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer panelServer.Close()

	p := New(&Config{NodesConfig: []*NodesConfig{
		{
			PanelType:        "NewV2board",
			ApiConfig:        &api.Config{APIHost: panelServer.URL, NodeID: 10, NodeType: "V2ray", EnableVless: true},
			ControllerConfig: &controller.Config{EnableFallback: true, DisableGetRule: true},
		},
		{
			PanelType:        "NewV2board",
			ApiConfig:        &api.Config{APIHost: panelServer.URL, NodeID: 11, NodeType: "V2ray", EnableVless: true},
			ControllerConfig: &controller.Config{ListenIP: "127.0.0.1", DisableGetRule: true},
		},
	}})
	if err := p.Start(); err != nil {
		t.Fatalf("panel should start with the healthy node: %v", err)
	}
	if !p.Running || p.Server == nil || len(p.Service) != 1 {
		t.Fatalf("expected one running node service, running=%v server=%v services=%d", p.Running, p.Server != nil, len(p.Service))
	}
	p.Close()
}

func TestStartNodeServicesSkipsFailedService(t *testing.T) {
	failed := &testService{startErr: errors.New("invalid node config")}
	healthy := &testService{}
	started, failures := startNodeServices([]nodeServiceEntry{
		{nodeID: 10, service: failed},
		{nodeID: 11, service: healthy},
	})

	if len(failures) != 1 || len(started) != 1 || started[0] != healthy {
		t.Fatalf("unexpected startup result: failures=%v started=%v", failures, started)
	}
	if failed.startCalls != 1 || failed.closeCalls != 1 {
		t.Fatalf("failed service was not cleaned up: starts=%d closes=%d", failed.startCalls, failed.closeCalls)
	}
	if healthy.startCalls != 1 || healthy.closeCalls != 0 {
		t.Fatalf("healthy service lifecycle is wrong: starts=%d closes=%d", healthy.startCalls, healthy.closeCalls)
	}
}

type testService struct {
	startErr   error
	startCalls int
	closeCalls int
}

func (s *testService) Start() error {
	s.startCalls++
	return s.startErr
}

func (s *testService) Close() error {
	s.closeCalls++
	return nil
}
