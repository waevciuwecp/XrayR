package controller

import (
	"github.com/XrayR-project/XrayR/api"
	"strings"
	"testing"
)

type invalidNodeAPI struct{ api.API }

func (invalidNodeAPI) Describe() api.ClientInfo { return api.ClientInfo{NodeID: 10, NodeType: "V2ray"} }
func (invalidNodeAPI) GetNodeInfo() (*api.NodeInfo, error) {
	return &api.NodeInfo{NodeType: "Vless", NodeID: 10, Port: 443, TransportProtocol: "tcp", Decryption: "mlkem768x25519plus.native.600s.synthetic"}, nil
}
func TestControllerStartRejectsFallback(t *testing.T) {
	c := &Controller{apiClient: invalidNodeAPI{}, config: &Config{EnableFallback: true}}
	if err := c.Start(); err == nil || !strings.Contains(err.Error(), "EnableFallback: false") {
		t.Fatalf("missing configuration error: %v", err)
	}
}
