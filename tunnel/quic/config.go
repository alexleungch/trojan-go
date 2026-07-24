package quic

import "github.com/p4gefau1t/trojan-go/config"

type HTTP3Config struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	KeyFile    string `json:"key" yaml:"key"`
	CertFile   string `json:"cert" yaml:"cert"`
	SNI        string `json:"sni" yaml:"sni"`
	Path       string `json:"path" yaml:"path"`
	Enable0RTT bool   `json:"enable_0rtt" yaml:"enable-0rtt"`
}

type Config struct {
	RemoteHost string      `json:"remote_addr" yaml:"remote-addr"`
	RemotePort int         `json:"remote_port" yaml:"remote-port"`
	LocalHost  string      `json:"local_addr" yaml:"local-addr"`
	LocalPort  int         `json:"local_port" yaml:"local-port"`
	HTTP3      HTTP3Config `json:"http3" yaml:"http3"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} { return new(Config) })
}
