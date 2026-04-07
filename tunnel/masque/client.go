package masque

import (
	"context"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Client struct {
	underlay tunnel.Client
	hostname string
	path     string
}

func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(nil, &Tunnel{})
	if err != nil {
		return nil, common.NewError("masque cannot dial with underlying client").Base(err)
	}

	log.Debug("masque connection established")
	return &OutboundConn{
		Conn:    conn,
		tcpConn: conn,
	}, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	return nil, common.NewError("not supported by masque")
}

func (c *Client) Close() error {
	return c.underlay.Close()
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	if cfg.Masque.Path == "" {
		return nil, common.NewError("masque path cannot be empty")
	}
	if cfg.Masque.Host == "" {
		cfg.Masque.Host = cfg.RemoteHost
		log.Warn("empty masque hostname")
	}
	log.Debug("masque client created")
	return &Client{
		hostname: cfg.Masque.Host,
		path:     cfg.Masque.Path,
		underlay: underlay,
	}, nil
}
