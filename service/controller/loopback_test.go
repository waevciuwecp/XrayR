package controller_test

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/app/mydispatcher"
	_ "github.com/XrayR-project/XrayR/cmd/distro/all"
	"github.com/XrayR-project/XrayR/common/mylego"
	. "github.com/XrayR-project/XrayR/service/controller"
	goreality "github.com/xtls/reality"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/routing"
	"github.com/xtls/xray-core/features/stats"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/proxy/vless"
	vlessin "github.com/xtls/xray-core/proxy/vless/inbound"
)

func loopbackCore(t *testing.T, raw string, inbound *core.InboundHandlerConfig) *core.Instance {
	t.Helper()
	var config conf.Config
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		t.Fatal(err)
	}
	built, err := config.Build()
	if err != nil {
		t.Fatal(err)
	}
	for i, app := range built.App {
		if app.Type == "xray.app.dispatcher.Config" {
			built.App[i] = serial.ToTypedMessage(&mydispatcher.Config{})
		}
	}
	if inbound != nil {
		built.Inbound = []*core.InboundHandlerConfig{inbound}
	}
	instance, err := core.New(built)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { instance.Close() })
	if _, ok := instance.GetFeature(routing.DispatcherType()).(*mydispatcher.DefaultDispatcher); !ok {
		t.Fatal("core is not using XrayR dispatcher")
	}
	if err := instance.Start(); err != nil {
		t.Fatal(err)
	}
	return instance
}

func TestVlessLoopback(t *testing.T) {
	// Every connection, TLS target and key is local to the test. No external DNS or services.
	echo, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { echo.Close() })
	go func() {
		for {
			conn, err := echo.Accept()
			if err != nil {
				return
			}
			go func() { defer conn.Close(); io.Copy(conn, conn) }()
		}
	}()
	largeTarget := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("local TLS largeTarget")) }))
	largeTarget.Config.ErrorLog = stdlog.New(io.Discard, "", 0)
	largeTarget.EnableHTTP2 = true
	certificateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certificate := &x509.Certificate{SerialNumber: big.NewInt(1), DNSNames: []string{"localhost"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, ExtraExtensions: []pkix.Extension{{Id: []int{1, 2, 3, 4}, Value: make([]byte, 4096)}}}
	der, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &certificateKey.PublicKey, certificateKey)
	if err != nil {
		t.Fatal(err)
	}
	// ML-DSA-65 needs room in the largeTarget certificate flight for its extra signature.
	largeTarget.TLS = &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: certificateKey}}}
	largeTarget.StartTLS()
	defer largeTarget.Close()
	target := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("local TLS target")) }))
	target.Config.ErrorLog = stdlog.New(io.Discard, "", 0)
	target.EnableHTTP2 = true
	target.TLS = &tls.Config{MinVersion: tls.VersionTLS13}
	target.StartTLS()
	defer target.Close()
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	private := base64.RawURLEncoding.EncodeToString(key.Bytes())
	public := base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes())
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		t.Fatal(err)
	}
	postQuantum, err := mldsa.NewPrivateKey(mldsa.MLDSA65(), seed)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, network                     string
		encrypted, reality, vision, mldsa bool
	}{
		{name: "tcp", network: "tcp"},
		{name: "tcp-encryption", network: "tcp", encrypted: true},
		{name: "vision-reality", network: "tcp", reality: true, vision: true},
		{name: "vision-reality-mldsa65", network: "tcp", reality: true, vision: true, mldsa: true},
		{name: "xhttp", network: "xhttp"},
		{name: "xhttp-reality-encryption", network: "xhttp", reality: true, encrypted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reservation, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			port := reservation.Addr().(*net.TCPAddr).Port
			reservation.Close()
			const id = "e5abf8da-5813-4b42-a328-e025d0433dd0"
			const email = "loopback|test@example.com|1"
			node := &api.NodeInfo{NodeType: "Vless", Port: uint32(port), TransportProtocol: tc.network, Path: "/loopback", XHTTPMode: "packet-up", EnableTLS: tc.reality, EnableREALITY: tc.reality}
			encryption := "none"
			flow := ""
			if tc.encrypted {
				node.Decryption = "mlkem768x25519plus.native.600s." + private
				encryption = "mlkem768x25519plus.native.0rtt." + public
			}
			if tc.vision {
				flow = "xtls-rprx-vision"
			}
			node.REALITYConfig = &api.REALITYConfig{Target: target.Listener.Addr().String(), ServerNames: []string{"localhost"}, PrivateKey: private, ShortIds: []string{"abcd"}}
			if tc.mldsa {
				node.REALITYConfig.Mldsa65Seed = base64.RawURLEncoding.EncodeToString(seed)
				node.REALITYConfig.Target = largeTarget.Listener.Addr().String()
			}
			inbound, err := InboundBuilder(&Config{ListenIP: "127.0.0.1", DisableLocalREALITYConfig: true, CertConfig: &mylego.CertConfig{CertMode: "none"}}, node, "loopback")
			if err != nil {
				t.Fatal(err)
			}
			message, err := inbound.ProxySettings.GetInstance()
			if err != nil {
				t.Fatal(err)
			}
			settings := message.(*vlessin.Config)
			settings.Users = []*protocol.User{{Email: email, Account: serial.ToTypedMessage(&vless.Account{Id: id, Flow: flow})}}
			inbound.ProxySettings = serial.ToTypedMessage(settings)
			server := loopbackCore(t, `{"log":{"loglevel":"none"},"policy":{"levels":{"0":{"statsUserUplink":true,"statsUserDownlink":true,"statsUserOnline":true}}},"stats":{},"outbounds":[{"protocol":"freedom","settings":{"finalRules":[{"action":"allow","ip":["127.0.0.0/8"]}]}}]}`, inbound)
			if tc.reality {
				// The core probes targets asynchronously; do not race its five-second fallback poll.
				probeKey := node.REALITYConfig.Target + " localhost 2"
				deadline := time.Now().Add(5 * time.Second)
				for {
					value, _ := goreality.GlobalPostHandshakeRecordsLens.Load(probeKey)
					if _, ready := value.([]int); ready {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("local REALITY target probe did not finish")
					}
					time.Sleep(time.Millisecond)
				}
			}
			dispatcher := server.GetFeature(mydispatcher.Type()).(*mydispatcher.DefaultDispatcher)
			users := []api.UserInfo{{UID: 1, Email: "test@example.com", UUID: id, DeviceLimit: 1}}
			if err := dispatcher.Limiter.AddInboundLimiter("loopback", 0, &users, nil); err != nil {
				t.Fatal(err)
			}
			stream := map[string]any{"network": tc.network}
			if tc.network == "xhttp" {
				stream["xhttpSettings"] = map[string]any{"path": "/loopback", "mode": "packet-up"}
			}
			if tc.reality {
				stream["security"] = "reality"
				stream["realitySettings"] = map[string]any{"serverName": "localhost", "fingerprint": "chrome", "password": public, "shortId": "abcd"}
				if tc.mldsa {
					stream["realitySettings"].(map[string]any)["mldsa65Verify"] = base64.RawURLEncoding.EncodeToString(postQuantum.PublicKey().Bytes())
				}
			}
			outbound := map[string]any{"protocol": "vless", "settings": map[string]any{"vnext": []any{map[string]any{"address": "127.0.0.1", "port": port, "users": []any{map[string]any{"id": id, "encryption": encryption, "flow": flow}}}}}, "streamSettings": stream}
			clientRaw, err := json.Marshal(map[string]any{"log": map[string]any{"loglevel": "none"}, "outbounds": []any{outbound}})
			if err != nil {
				t.Fatal(err)
			}
			client := loopbackCore(t, string(clientRaw), nil)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			destination, err := xnet.ParseDestination("tcp:" + echo.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			conn, err := core.Dial(ctx, client, destination)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			payload := []byte("XrayR loopback " + tc.name)
			done := make(chan error, 1)
			go func() {
				if _, err := conn.Write(payload); err != nil {
					done <- err
					return
				}
				got := make([]byte, len(payload))
				_, err := io.ReadFull(conn, got)
				if err == nil && string(got) != string(payload) {
					err = fmt.Errorf("echo mismatch")
				}
				done <- err
			}()
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				conn.Close()
				t.Fatal(ctx.Err())
			}
			manager := server.GetFeature(stats.ManagerType()).(stats.Manager)
			for _, direction := range []string{"uplink", "downlink"} {
				counter := manager.GetCounter("user>>>" + email + ">>>traffic>>>" + direction)
				if counter == nil || counter.Value() != int64(len(payload)) {
					t.Fatalf("%s accounting: %v, want %d", direction, counter, len(payload))
				}
			}
			online := manager.GetOnlineMap("user>>>" + email + ">>>online")
			if online == nil || online.Count() != 0 {
				t.Fatal("online map missing or loopback exclusion changed")
			}
			devices, err := dispatcher.Limiter.GetOnlineDevice("loopback")
			if err != nil || len(*devices) != 1 {
				t.Fatalf("device accounting missing: %v", err)
			}
		})
	}
}
