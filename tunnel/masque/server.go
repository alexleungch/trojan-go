package masque

import (
	"context"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/redirector"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Server struct {
	underlay  tunnel.Server
	hostname  string
	path      string
	enabled   bool
	redirAddr net.Addr
	redir     *redirector.Redirector
	ctx       context.Context
	cancel    context.CancelFunc
	timeout   time.Duration
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := s.underlay.AcceptConn(&Tunnel{})
	if err != nil {
		return nil, common.NewError("masque failed to accept connection from underlying server")
	}

	if !s.enabled {
		s.redir.Redirect(&redirector.Redirection{
			InboundConn: conn,
			RedirectTo:  s.redirAddr,
		})
		return nil, common.NewError("masque is disabled. redirecting http request from " + conn.RemoteAddr().String())
	}

	log.Debug("masque connection accepted")
	ctx, cancel := context.WithCancel(s.ctx)

	return &InboundConn{
		OutboundConn: OutboundConn{
			Conn:    conn,
			tcpConn: conn,
		},
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	return nil, common.NewError("not supported")
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	if cfg.Masque.Enabled {
		if cfg.Masque.Path == "" {
			return nil, common.NewError("masque path cannot be empty")
		}
	}
	if cfg.RemoteHost == "" {
		log.Warn("empty masque redirection hostname")
		cfg.RemoteHost = cfg.Masque.Host
	}
	if cfg.RemotePort == 0 {
		log.Warn("empty masque redirection port")
		cfg.RemotePort = 80
	}
	ctx, cancel := context.WithCancel(ctx)
	log.Debug("masque server created")
	return &Server{
		enabled:   cfg.Masque.Enabled,
		hostname:  cfg.Masque.Host,
		path:      cfg.Masque.Path,
		ctx:       ctx,
		cancel:    cancel,
		underlay:  underlay,
		timeout:   time.Second * time.Duration(10),
		redir:     redirector.NewRedirector(ctx),
		redirAddr: tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort),
	}, nil
}
