package controller

import (
	"encoding/json"
	"fmt"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"

	"github.com/XrayR-project/XrayR/api"
)

// OutboundBuilder build freedom or loopback outbound config for addOutbound
func OutboundBuilder(config *Config, nodeInfo *api.NodeInfo, tag string) (*core.OutboundHandlerConfig, error) {
	outboundDetourConfig := &conf.OutboundDetourConfig{}
	outboundDetourConfig.Tag = tag

	// SendThrough setting
	outboundDetourConfig.SendThrough = &config.SendIP

	// MuxConfig (multiplexing)
	if config.MuxEnabled {
		outboundDetourConfig.MuxSettings = &conf.MuxConfig{
			Enabled:         config.MuxEnabled,
			Concurrency:     config.MuxConcurrency,
			XudpConcurrency: config.MuxXudpConcurrency,
			XudpProxyUDP443: config.MuxXudpProxyUDP443,
		}
	}

	if nodeInfo.NodeType == "Loopback" {
		outboundDetourConfig.Protocol = "loopback"
		proxySetting := &conf.LoopbackConfig{
			InboundTag: tag,
		}
		var setting json.RawMessage
		setting, err := json.Marshal(proxySetting)
		if err != nil {
			return nil, fmt.Errorf("marshal loopback config failed: %s", err)
		}
		outboundDetourConfig.Settings = &setting
		return outboundDetourConfig.Build()
	}

	outboundDetourConfig.Protocol = "freedom"

	// Freedom Protocol setting
	var domainStrategy = "Asis"
	if config.EnableDNS {
		if config.DNSType != "" {
			domainStrategy = config.DNSType
		} else {
			domainStrategy = "UseIP"
		}
	}
	proxySetting := &conf.FreedomConfig{
		DomainStrategy: domainStrategy,
	}
	// TCP Fragment for anti-censorship
	if config.FragmentPackets != "" {
		proxySetting.Fragment = &conf.Fragment{
			Packets: config.FragmentPackets,
		}
		if config.FragmentLength != nil {
			proxySetting.Fragment.Length = &conf.Int32Range{
				From: config.FragmentLength.From,
				To:   config.FragmentLength.To,
			}
		}
		if config.FragmentInterval != nil {
			proxySetting.Fragment.Interval = &conf.Int32Range{
				From: config.FragmentInterval.From,
				To:   config.FragmentInterval.To,
			}
		}
		if config.FragmentMaxSplit != nil {
			proxySetting.Fragment.MaxSplit = &conf.Int32Range{
				From: config.FragmentMaxSplit.From,
				To:   config.FragmentMaxSplit.To,
			}
		}
	}
	// Used for Shadowsocks-Plugin
	if nodeInfo.NodeType == "dokodemo-door" {
		proxySetting.Redirect = fmt.Sprintf("127.0.0.1:%d", nodeInfo.Port-1)
	}
	var setting json.RawMessage
	setting, err := json.Marshal(proxySetting)
	if err != nil {
		return nil, fmt.Errorf("marshal proxy %s config failed: %s", nodeInfo.NodeType, err)
	}
	outboundDetourConfig.Settings = &setting
	return outboundDetourConfig.Build()
}
