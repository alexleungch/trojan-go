package quic

import "github.com/p4gefau1t/trojan-go/config"

type QuicConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	KeyFile    string `json:"key" yaml:"key"`
	CertFile   string `json:"cert" yaml:"cert"`
	SNI        string `json:"sni" yaml:"sni"`
	ALPN       string `json:"alpn" yaml:"alpn"`
	InitialRTT string `json:"initial_rtt" yaml:"initial-rtt"`
	MaxIdleTO  string `json:"max_idle_timeout" yaml:"max-idle-timeout"`
}

type Config struct {
	RemoteHost string     `json:"remote_addr" yaml:"remote-addr"`
	RemotePort int        `json:"remote_port" yaml:"remote-port"`
	LocalHost  string     `json:"local_addr" yaml:"local-addr"`
	LocalPort  int        `json:"local_port" yaml:"local-port"`
	Quic       QuicConfig `json:"quic" yaml:"quic"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}

