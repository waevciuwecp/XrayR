package newV2board

import (
	"encoding/json"
	"github.com/xtls/xray-core/infra/conf"
)

type serverConfig struct {
	shadowsocks
	v2ray
	trojan

	ServerPort int `json:"server_port"`
	BaseConfig struct {
		PushInterval int `json:"push_interval"`
		PullInterval int `json:"pull_interval"`
	} `json:"base_config"`
	Routes []route `json:"routes"`
}

type shadowsocks struct {
	Cipher       string `json:"cipher"`
	Obfs         string `json:"obfs"`
	ObfsSettings struct {
		Path string `json:"path"`
		Host string `json:"host"`
	} `json:"obfs_settings"`
	ServerKey string `json:"server_key"`
}

type v2ray struct {
	Network              string           `json:"network"`
	NetworkSettings      networkSettings  `json:"networkSettings"`
	VlessNetworkSettings *networkSettings `json:"network_settings"`
	Decryption           string           `json:"decryption"`
	VlessFlow            string           `json:"flow"`
	VlessTlsSettings     struct {
		Target                string             `json:"target"`
		Mldsa65Seed           string             `json:"mldsa65_seed"`
		LimitFallbackUpload   conf.LimitFallback `json:"limit_fallback_upload"`
		LimitFallbackDownload conf.LimitFallback `json:"limit_fallback_download"`
		ServerPort            string             `json:"server_port"`
		Dest                  string             `json:"dest"`
		XVer                  uint64             `json:"xver"`
		Sni                   string             `json:"server_name"`
		PrivateKey            string             `json:"private_key"`
		ShortId               string             `json:"short_id"`
	} `json:"tls_settings"`
	Tls int `json:"tls"`
}

type trojan struct {
	Host       string `json:"host"`
	ServerName string `json:"server_name"`
}

type route struct {
	Id          int      `json:"id"`
	Match       []string `json:"match"`
	Action      string   `json:"action"`
	ActionValue string   `json:"action_value"`
}

type user struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	SpeedLimit int    `json:"speed_limit"`
}

type networkSettings struct {
	Mode        string           `json:"mode"`
	Extra       json.RawMessage  `json:"extra"`
	Path        string           `json:"path"`
	Host        string           `json:"host"`
	Headers     *json.RawMessage `json:"headers"`
	ServiceName string           `json:"serviceName"`
	Header      *json.RawMessage `json:"header"`
}
