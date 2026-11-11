package controller_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/common/mylego"
	. "github.com/XrayR-project/XrayR/service/controller"
	"github.com/xtls/xray-core/app/proxyman"
	xraytls "github.com/xtls/xray-core/transport/internet/tls"
)

func TestHysteria2RequiresTLSCertificate(t *testing.T) {
	node := &api.NodeInfo{NodeType: "Hysteria", NodeID: 14, Port: 443, TransportProtocol: "hysteria", EnableTLS: true}
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name:   "missing certificate even when REALITY is globally enabled",
			config: &Config{EnableREALITY: true, REALITYConfigs: &REALITYConfig{}},
			want:   "TLS requires CertConfig",
		},
		{
			name:   "certificate mode none",
			config: &Config{CertConfig: &mylego.CertConfig{CertMode: "none"}},
			want:   "Hysteria 2 requires TLS certificates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InboundBuilder(tt.config, node, "hy2")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got error %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestHysteria2RejectsREALITY(t *testing.T) {
	node := &api.NodeInfo{NodeType: "Hysteria", NodeID: 14, Port: 443, TransportProtocol: "hysteria", EnableTLS: true, EnableREALITY: true}
	_, err := InboundBuilder(&Config{DisableLocalREALITYConfig: true}, node, "hy2")
	if err == nil || !strings.Contains(err.Error(), "Hysteria 2 does not support REALITY") {
		t.Fatalf("got error %v, want explicit Hysteria 2 REALITY rejection", err)
	}
}

func TestHysteria2UsesH3ALPN(t *testing.T) {
	node := &api.NodeInfo{NodeType: "Hysteria", NodeID: 14, Port: 443, TransportProtocol: "hysteria", EnableTLS: true}
	certFile, keyFile := hysteriaTestCertificate(t)
	for _, configuredALPN := range []string{"", "h2"} {
		t.Run("configured="+configuredALPN, func(t *testing.T) {
			inbound, err := InboundBuilder(&Config{
				EnableREALITY:  true,
				REALITYConfigs: &REALITYConfig{},
				TLSAlpn:        configuredALPN,
				CertConfig: &mylego.CertConfig{
					CertMode: "file",
					CertFile: certFile,
					KeyFile:  keyFile,
				},
			}, node, "hy2")
			if err != nil {
				t.Fatal(err)
			}

			receiverMessage, err := inbound.ReceiverSettings.GetInstance()
			if err != nil {
				t.Fatal(err)
			}
			receiver, ok := receiverMessage.(*proxyman.ReceiverConfig)
			if !ok {
				t.Fatalf("unexpected receiver settings type %T", receiverMessage)
			}
			streamSettings := receiver.GetStreamSettings()
			if streamSettings.GetProtocolName() != "hysteria" || !strings.HasSuffix(streamSettings.GetSecurityType(), ".tls.Config") {
				t.Fatalf("Hysteria 2 did not build a TLS Hysteria stream: %+v", streamSettings)
			}
			securitySettings := streamSettings.GetSecuritySettings()
			if len(securitySettings) != 1 {
				t.Fatalf("unexpected TLS settings count: %d", len(securitySettings))
			}
			tlsMessage, err := securitySettings[0].GetInstance()
			if err != nil {
				t.Fatal(err)
			}
			tlsSettings, ok := tlsMessage.(*xraytls.Config)
			if !ok {
				t.Fatalf("unexpected TLS settings type %T", tlsMessage)
			}
			if len(tlsSettings.NextProtocol) != 1 || tlsSettings.NextProtocol[0] != "h3" {
				t.Fatalf("Hysteria 2 must advertise only h3 ALPN: %v", tlsSettings.NextProtocol)
			}
		})
	}
}

func hysteriaTestCertificate(t *testing.T) (string, string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"hy2.example.com"},
	}
	certificate, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate}), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}), 0o600); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}
