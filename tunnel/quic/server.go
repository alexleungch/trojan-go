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

type Server struct {
listener   quic.Listener
underlay   tunnel.Server
ctx        context.Context
cancel     context.CancelFunc
serverName string
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
session, err := s.listener.Accept(s.ctx)
if err != nil {
return nil, common.NewError("quic failed to accept session").Base(err)
}

stream, err := session.AcceptStream(s.ctx)
if err != nil {
session.CloseWithError(0, "failed to accept stream")
return nil, common.NewError("quic failed to accept stream").Base(err)
}

log.Debug("quic stream accepted")
return &Conn{
Stream: stream,
conn:   session,
}, nil
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
return nil, common.NewError("not supported by quic")
}

func (s *Server) Close() error {
s.cancel()
return s.listener.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
cfg := config.FromContext(ctx, Name).(*Config)

if !cfg.Quic.Enabled {
return nil, common.NewError("quic is not enabled")
}

serverName := cfg.Quic.SNI
if serverName == "" {
serverName = cfg.LocalHost
}

cert, err := tls.LoadX509KeyPair(cfg.Quic.CertFile, cfg.Quic.KeyFile)
if err != nil {
return nil, common.NewError("quic failed to load cert/key").Base(err)
}

tlsConfig := &tls.Config{
Certificates: []tls.Certificate{cert},
NextProtos:   []string{cfg.Quic.ALPN},
}

if cfg.Quic.ALPN == "" {
tlsConfig.NextProtos = []string{"http/3"}
}

listenAddr := tunnel.NewAddressFromHostPort("udp", cfg.LocalHost, cfg.LocalPort)
listener, err := quic.ListenAddr(listenAddr.String(), tlsConfig, &quic.Config{})
if err != nil {
return nil, common.NewError("quic failed to listen").Base(err)
}

ctx, cancel := context.WithCancel(ctx)
log.Debug("quic server created")
return &Server{
listener:   listener,
underlay:   underlay,
ctx:        ctx,
cancel:     cancel,
serverName: serverName,
}, nil
}
