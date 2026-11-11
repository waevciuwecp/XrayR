package controller

import (
	"reflect"
	"strings"

	"github.com/XrayR-project/XrayR/api"
)

func (c *Controller) logPanelMultiplex(previous, current *api.NodeInfo) {
	if current == nil {
		return
	}

	var previousMultiplex *api.MultiplexConfig
	if previous != nil {
		previousMultiplex = previous.Multiplex
	}
	if reflect.DeepEqual(previousMultiplex, current.Multiplex) {
		return
	}

	if current.Multiplex == nil || !current.Multiplex.Enabled {
		if previousMultiplex != nil && previousMultiplex.Enabled {
			c.logger.Infof("Panel multiplex disabled for node %d (%s)", current.NodeID, current.NodeType)
		}
		return
	}

	switch strings.ToLower(current.NodeType) {
	case "vless", "vmess", "v2ray":
		c.logger.Infof("Panel multiplex enabled for node %d (%s, protocol=%q); this is a client subscription setting, not an XrayR inbound setting. Xray-core accepts only compatible native mux traffic.", current.NodeID, current.NodeType, current.Multiplex.Protocol)
	case "trojan":
		c.logger.Warnf("Panel multiplex enabled for Trojan node %d (protocol=%q); XrayR does not apply panel mux settings, and compatibility with Xray-core native mux is not guaranteed.", current.NodeID, current.Multiplex.Protocol)
	default:
		c.logger.Warnf("Panel multiplex enabled for node %d (%s, protocol=%q); XrayR does not translate it into server-side mux configuration.", current.NodeID, current.NodeType, current.Multiplex.Protocol)
	}
}

func (c *Controller) warnLegacyOutboundMux() {
	if c.config.MuxEnabled {
		c.logger.Warn("MuxEnabled is ignored for XrayR-generated freedom/loopback outbounds; it does not control panel client multiplex settings")
	}
}
