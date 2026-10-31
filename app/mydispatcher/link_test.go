package mydispatcher

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/common/limiter"
	appstats "github.com/xtls/xray-core/app/stats"
	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/features/policy"
	"github.com/xtls/xray-core/transport"
)

type linkPolicy struct{ policy.Manager }

func (linkPolicy) ForLevel(uint32) policy.Session {
	p := policy.SessionDefault()
	p.Stats.UserUplink = true
	p.Stats.UserDownlink = true
	p.Stats.UserOnline = true
	return p
}

func TestDispatcherLinkAccounting(t *testing.T) {
	manager, err := appstats.NewManager(context.Background(), &appstats.Config{})
	if err != nil {
		t.Fatal(err)
	}
	dispatcher := &DefaultDispatcher{policy: linkPolicy{}, stats: manager, Limiter: limiter.New()}
	users := []api.UserInfo{{UID: 1, Email: "test@example.com", SpeedLimit: 128, DeviceLimit: 1}}
	if err := dispatcher.Limiter.AddInboundLimiter("test", 0, &users, nil); err != nil {
		t.Fatal(err)
	}
	const email = "test|test@example.com|1"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	makeContext := func(ip string) context.Context {
		return session.ContextWithInbound(ctx, &session.Inbound{Tag: "test", User: &protocol.MemoryUser{Email: email}, Source: net.TCPDestination(net.ParseAddress(ip), 1234)})
	}
	link := &transport.Link{Reader: buf.NewReader(strings.NewReader("uplink")), Writer: buf.Discard}
	if err := dispatcher.wrapLink(makeContext("192.0.2.1"), link); err != nil {
		t.Fatal(err)
	}
	mb, err := link.Reader.ReadMultiBuffer()
	if err != nil {
		t.Fatal(err)
	}
	buf.ReleaseMulti(mb)
	if err := link.Writer.WriteMultiBuffer(buf.MergeBytes(nil, []byte("downlink"))); err != nil {
		t.Fatal(err)
	}
	for direction, want := range map[string]int64{"uplink": 6, "downlink": 8} {
		c := manager.GetCounter("user>>>" + email + ">>>traffic>>>" + direction)
		if c == nil || c.Value() != want {
			t.Fatalf("%s counter=%v, want %d", direction, c, want)
		}
	}
	online := manager.GetOnlineMap("user>>>" + email + ">>>online")
	if online == nil || online.Count() != 1 {
		t.Fatal("online tracking missing")
	}
	second := &transport.Link{Reader: buf.NewReader(strings.NewReader("")), Writer: buf.Discard}
	if err := dispatcher.wrapLink(makeContext("192.0.2.2"), second); err == nil {
		t.Fatal("second device was not rejected")
	}
	cancel()
	deadline := time.Now().Add(time.Second)
	for online.Count() != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if online.Count() != 0 {
		t.Fatal("online entry survived cancellation")
	}
}
