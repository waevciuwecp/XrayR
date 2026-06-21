package controller

import (
	"github.com/XrayR-project/XrayR/common/limiter"
	"github.com/XrayR-project/XrayR/common/mylego"
)

type Config struct {
	ListenIP                  string                           `mapstructure:"ListenIP"`
	SendIP                    string                           `mapstructure:"SendIP"`
	UpdatePeriodic            int                              `mapstructure:"UpdatePeriodic"`
	CertConfig                *mylego.CertConfig               `mapstructure:"CertConfig"`
	EnableDNS                 bool                             `mapstructure:"EnableDNS"`
	DNSType                   string                           `mapstructure:"DNSType"`
	DisableUploadTraffic      bool                             `mapstructure:"DisableUploadTraffic"`
	DisableGetRule            bool                             `mapstructure:"DisableGetRule"`
	EnableProxyProtocol       bool                             `mapstructure:"EnableProxyProtocol"`
	EnableFallback            bool                             `mapstructure:"EnableFallback"`
	DisableIVCheck            bool                             `mapstructure:"DisableIVCheck"`
	DisableSniffing           bool                             `mapstructure:"DisableSniffing"`
	AutoSpeedLimitConfig      *AutoSpeedLimitConfig            `mapstructure:"AutoSpeedLimitConfig"`
	GlobalDeviceLimitConfig   *limiter.GlobalDeviceLimitConfig `mapstructure:"GlobalDeviceLimitConfig"`
	FallBackConfigs           []*FallBackConfig                `mapstructure:"FallBackConfigs"`
	DisableLocalREALITYConfig bool                             `mapstructure:"DisableLocalREALITYConfig"`
	EnableREALITY             bool                             `mapstructure:"EnableREALITY"`
	REALITYConfigs            *REALITYConfig                   `mapstructure:"REALITYConfigs"`

	// SniffingConfig extended fields
	SniffingDomainsExcluded []string `mapstructure:"SniffingDomainsExcluded"`
	SniffingMetadataOnly    bool     `mapstructure:"SniffingMetadataOnly"`
	SniffingRouteOnly       bool     `mapstructure:"SniffingRouteOnly"`

	// TLS extended fields
	TLSFingerprint string `mapstructure:"TLSFingerprint"`
	TLSAlpn        string `mapstructure:"TLSAlpn"`

	// REALITY extended fields
	REALITYFingerprint string `mapstructure:"REALITYFingerprint"`
	REALITYSpiderX     string `mapstructure:"REALITYSpiderX"`
	REALITYPublicKey   string `mapstructure:"REALITYPublicKey"`
	REALITYShortId     string `mapstructure:"REALITYShortId"`

	// SocketConfig extended fields
	SOMark                uint32   `mapstructure:"SOMark"`
	TCPFastOpen           bool     `mapstructure:"TCPFastOpen"`
	TProxy                string   `mapstructure:"TProxy"`
	DialerProxy           string   `mapstructure:"DialerProxy"`
	BindInterface         string   `mapstructure:"BindInterface"`
	TCPKeepAliveInterval  int32    `mapstructure:"TCPKeepAliveInterval"`
	TCPCongestion         string   `mapstructure:"TCPCongestion"`
	TrustedXForwardedFor  []string `mapstructure:"TrustedXForwardedFor"`

	// XHTTP/SplitHTTP extended fields
	XHTTPMode           string      `mapstructure:"XHTTPMode"`
	XHTTPNoGRPCHeader   bool        `mapstructure:"XHTTPNoGRPCHeader"`
	XHTTPNoSSEHeader    bool        `mapstructure:"XHTTPNoSSEHeader"`
	XHTTPPaddingBytes   *Int32Range `mapstructure:"XHTTPPaddingBytes"`

	// WebSocket extended fields
	WSHeartbeatPeriod uint32 `mapstructure:"WSHeartbeatPeriod"`

	// GRPC extended fields
	GRPCMultiMode          bool   `mapstructure:"GRPCMultiMode"`
	GRPCIdleTimeout        int32  `mapstructure:"GRPCIdleTimeout"`
	GRPCHealthCheckTimeout int32  `mapstructure:"GRPCHealthCheckTimeout"`
	GRPCPermitWithoutStream bool  `mapstructure:"GRPCPermitWithoutStream"`
	GRPCInitialWindowsSize int32  `mapstructure:"GRPCInitialWindowsSize"`
	GRPCUserAgent          string `mapstructure:"GRPCUserAgent"`

	// FinalMask JSON config (raw JSON matching Xray-core FinalMask structure)
	FinalMaskJSON string `mapstructure:"FinalMaskJSON"`

	// MuxConfig (multiplexing)
	MuxEnabled         bool   `mapstructure:"MuxEnabled"`
	MuxConcurrency     int16  `mapstructure:"MuxConcurrency"`
	MuxXudpConcurrency int16  `mapstructure:"MuxXudpConcurrency"`
	MuxXudpProxyUDP443 string `mapstructure:"MuxXudpProxyUDP443"`

	// Freedom Fragment (TCP fragmentation for anti-censorship)
	FragmentPackets  string      `mapstructure:"FragmentPackets"`
	FragmentLength   *Int32Range `mapstructure:"FragmentLength"`
	FragmentInterval *Int32Range `mapstructure:"FragmentInterval"`
	FragmentMaxSplit *Int32Range `mapstructure:"FragmentMaxSplit"`
}

type Int32Range struct {
	From int32 `mapstructure:"From"`
	To   int32 `mapstructure:"To"`
}

type AutoSpeedLimitConfig struct {
	Limit         int `mapstructure:"Limit"` // mbps
	WarnTimes     int `mapstructure:"WarnTimes"`
	LimitSpeed    int `mapstructure:"LimitSpeed"`    // mbps
	LimitDuration int `mapstructure:"LimitDuration"` // minute
}

type FallBackConfig struct {
	SNI              string `mapstructure:"SNI"`
	Alpn             string `mapstructure:"Alpn"`
	Path             string `mapstructure:"Path"`
	Dest             string `mapstructure:"Dest"`
	ProxyProtocolVer uint64 `mapstructure:"ProxyProtocolVer"`
}

type REALITYConfig struct {
	Show             bool     `mapstructure:"Show"`
	Dest             string   `mapstructure:"Dest"`
	ProxyProtocolVer uint64   `mapstructure:"ProxyProtocolVer"`
	ServerNames      []string `mapstructure:"ServerNames"`
	PrivateKey       string   `mapstructure:"PrivateKey"`
	MinClientVer     string   `mapstructure:"MinClientVer"`
	MaxClientVer     string   `mapstructure:"MaxClientVer"`
	MaxTimeDiff      uint64   `mapstructure:"MaxTimeDiff"`
	ShortIds         []string `mapstructure:"ShortIds"`
}
