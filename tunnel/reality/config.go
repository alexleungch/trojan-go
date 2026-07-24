package reality

import (
	"github.com/p4gefau1t/trojan-go/config"
)

type Config struct {
	Reality RealityConfig `json:"reality" yaml:"reality"`
}

type RealityConfig struct {
	Enabled     bool     `json:"enabled" yaml:"enabled"`
	Fingerprint string   `json:"fingerprint" yaml:"fingerprint"`
	ServerName  string   `json:"server_name" yaml:"server-name"`
	PublicKey   string   `json:"public_key" yaml:"public-key"`
	ShortID     string   `json:"short_id" yaml:"short-id"`

	// Server-side settings
	Dest        string   `json:"dest" yaml:"dest"`
	ServerNames []string `json:"server_names" yaml:"server-names"`
	PrivateKey  string   `json:"private_key" yaml:"private-key"`
	ShortIDs    []string `json:"short_ids" yaml:"short-ids"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			Reality: RealityConfig{
				Fingerprint: "chrome",
			},
		}
	})
}
