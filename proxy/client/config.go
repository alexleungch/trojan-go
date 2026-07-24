package client

import "github.com/p4gefau1t/trojan-go/config"

type MuxConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}
type WebsocketConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}
type RouterConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}
type RealityConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}
type HTTP3Config struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}
type TransportPluginConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type Config struct {
	Mux             MuxConfig             `json:"mux" yaml:"mux"`
	Websocket       WebsocketConfig       `json:"websocket" yaml:"websocket"`
	Router          RouterConfig          `json:"router" yaml:"router"`
	Reality         RealityConfig          `json:"reality" yaml:"reality"`
	HTTP3           HTTP3Config           `json:"http3" yaml:"http3"`
	TransportPlugin TransportPluginConfig `json:"transport_plugin" yaml:"transport-plugin"`
}

func init() { config.RegisterConfigCreator(Name, func() interface{} { return new(Config) }) }
