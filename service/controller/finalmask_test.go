package controller

import (
	"encoding/json"
	"testing"

	"github.com/XrayR-project/XrayR/api"
)

func TestBuildFinalMaskMergesHysteria2Salamander(t *testing.T) {
	config := &Config{FinalMaskJSON: `{"udp":[{"type":"salamander","settings":{"password":"global-password"}},{"type":"filter","settings":{}}]}`}
	nodeInfo := &api.NodeInfo{
		NodeType:     "Hysteria",
		NodeID:       17,
		ObfsType:     "salamander",
		ObfsPassword: "panel-password",
	}

	finalMask, err := buildFinalMask(config, nodeInfo)
	if err != nil {
		t.Fatalf("buildFinalMask: %v", err)
	}
	if got, want := len(finalMask.Udp), 3; got != want {
		t.Fatalf("got %d UDP masks, want %d", got, want)
	}
	if got, want := finalMask.Udp[0].Type, "salamander"; got != want {
		t.Fatalf("global mask moved: got first type %q, want %q", got, want)
	}
	if got, want := finalMask.Udp[1].Type, "salamander"; got != want {
		t.Fatalf("got node mask type %q, want %q", got, want)
	}
	if got, want := finalMask.Udp[2].Type, "filter"; got != want {
		t.Fatalf("filter must remain outermost: got final type %q, want %q", got, want)
	}
	var settings struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(*finalMask.Udp[1].Settings, &settings); err != nil {
		t.Fatalf("decode node mask settings: %v", err)
	}
	if got, want := settings.Password, nodeInfo.ObfsPassword; got != want {
		t.Fatalf("got node obfs password %q, want %q", got, want)
	}
}

func TestBuildFinalMaskDoesNotEnableStaleHysteria2Password(t *testing.T) {
	finalMask, err := buildFinalMask(&Config{}, &api.NodeInfo{
		NodeType:     "Hysteria",
		ObfsPassword: "stale-password",
	})
	if err != nil {
		t.Fatalf("buildFinalMask: %v", err)
	}
	if finalMask != nil {
		t.Fatalf("got final mask %#v, want nil", finalMask)
	}
}
