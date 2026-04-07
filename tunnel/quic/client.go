package quic

import (
"context"
"crypto/tls"

"github.com/quic-go/quic-go"

"github.com/p4gefau1t/trojan-go/common"
"github.com/p4gefau1t/trojan-go/config"
"github.com/p4gefau1t/trojan-go/log"
"github.com/p4gefau1t/trojan-go/tunnel"
)

type Client struct {
session    quic.Connection
underlay   tunnel.Client
serverName string
tlsConfig  *tls.Config
}

func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
stream, err := c.session.OpenStreamSync(context.Background())
if err != nil {
return nil, common.NewError("quic failed to open stream").Base(err)
}
log.Debug("quic stream opened")
return &Conn{
Stream: stream,
conn:   c.session,
}, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
return nil, common.NewError("not supported by quic")
}

func (c *Client) Close() error {
if c.session != nil {
c.session.CloseWithError(0, "client closed")
}
return c.underlay.Close()
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
cfg := config.FromContext(ctx, Name).(*Config)

serverName := cfg.Quic.SNI
if serverName == "" {
serverName = cfg.RemoteHost
}

tlsConfig := &tls.Config{
ServerName:         serverName,
InsecureSkipVerify: false,
NextProtos:         []string{cfg.Quic.ALPN},
}

if cfg.Quic.ALPN == "" {
tlsConfig.NextProtos = []string{"http/3"}
}

quicConfig := &quic.Config{
MaxIdleTimeout: 0,
}

session, err := quic.DialAddr(ctx,
tunnel.NewAddressFromHostPort("udp", cfg.RemoteHost, cfg.RemotePort).String(),
tlsConfig,
quicConfig)
if err != nil {
return nil, common.NewError("quic failed to dial").Base(err)
}

log.Debug("quic client created")
return &Client{
session:    session,
underlay:   underlay,
serverName: serverName,
tlsConfig:  tlsConfig,
}, nil
}
