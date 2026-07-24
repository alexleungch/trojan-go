package reality

import (
	"context"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"

	reality "github.com/xtls/reality"
	utls "github.com/refraction-networking/utls"
)

// Client is a REALITY client
type Client struct {
	serverName  string
	fingerprint string
	helloID     utls.ClientHelloID
	underlay    tunnel.Client
}

func (c *Client) Close() error {
	return c.underlay.Close()
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (c *Client) DialConn(_ *tunnel.Address, overlay tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(nil, &Tunnel{})
	if err != nil {
		return nil, common.NewError("reality failed to dial conn").Base(err)
	}

	if c.fingerprint != "" {
		tlsConn := utls.UClient(conn, &utls.Config{
			ServerName:         c.serverName,
			InsecureSkipVerify: true,
		}, c.helloID)
		if err := tlsConn.Handshake(); err != nil {
			return nil, common.NewError("reality failed to handshake with remote server").Base(err)
		}
		return &transport.Conn{
			Conn: tlsConn,
		}, nil
	}

	realityConfig := &reality.Config{
		ServerName:         c.serverName,
		InsecureSkipVerify: true,
	}
	realityConn := reality.Client(conn, realityConfig)
	if err := realityConn.Handshake(); err != nil {
		return nil, common.NewError("reality failed to handshake with remote server").Base(err)
	}
	return &Conn{
		Conn: realityConn,
	}, nil
}

// NewClient creates a REALITY client
func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)

	helloID := utls.ClientHelloID{}
	if cfg.Reality.Fingerprint != "" {
		switch cfg.Reality.Fingerprint {
		case "firefox":
			helloID = utls.HelloFirefox_Auto
		case "chrome":
			helloID = utls.HelloChrome_Auto
		case "ios":
			helloID = utls.HelloIOS_Auto
		default:
			return nil, common.NewError("invalid fingerprint " + cfg.Reality.Fingerprint)
		}
		log.Info("reality fingerprint", cfg.Reality.Fingerprint, "applied")
	}

	if cfg.Reality.ServerName == "" {
		return nil, common.NewError("reality server_name is required")
	}

	client := &Client{
		underlay:    underlay,
		serverName:  cfg.Reality.ServerName,
		fingerprint: cfg.Reality.Fingerprint,
		helloID:     helloID,
	}

	log.Debug("reality client created")
	return client, nil
}
