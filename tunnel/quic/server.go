package quic

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync"

	quicgo "github.com/quic-go/quic-go"
	h3 "github.com/quic-go/quic-go/http3"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Server struct {
	server    *h3.Server
	packet    net.PacketConn
	connChan  chan tunnel.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("http3 server closed")
	}
}
func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	return nil, common.NewError("http3 does not support packet tunnels")
}
func (s *Server) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.cancel()
		err = s.server.Close()
		if closeErr := s.packet.Close(); err == nil {
			err = closeErr
		}
	})
	return err
}

func NewServer(ctx context.Context, _ tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	if !cfg.HTTP3.Enabled {
		return nil, common.NewError("http3 is not enabled")
	}
	cert, err := tls.LoadX509KeyPair(cfg.HTTP3.CertFile, cfg.HTTP3.KeyFile)
	if err != nil {
		return nil, common.NewError("http3 failed to load cert/key").Base(err)
	}
	path := cfg.HTTP3.Path
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		return nil, common.NewError("http3 path must start with /")
	}
	packet, err := net.ListenPacket("udp", tunnel.NewAddressFromHostPort("udp", cfg.LocalHost, cfg.LocalPort).String())
	if err != nil {
		return nil, common.NewError("http3 failed to listen").Base(err)
	}
	serverCtx, cancel := context.WithCancel(ctx)
	s := &Server{packet: packet, connChan: make(chan tunnel.Conn, 32), ctx: serverCtx, cancel: cancel}
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		remoteAddr, _ := net.ResolveUDPAddr("udp", r.RemoteAddr)
		conn := &Conn{reader: r.Body, writer: writeCloser{w}, localAddr: packet.LocalAddr(), remoteAddr: remoteAddr}
		select {
		case s.connChan <- conn:
			select {
			case <-serverCtx.Done():
			case <-r.Context().Done():
			}
		case <-serverCtx.Done():
		}
		conn.Close()
	})
	quicConfig := &quicgo.Config{}
	if cfg.HTTP3.Enable0RTT {
		quicConfig.Allow0RTT = true
	}
	s.server = &h3.Server{Handler: mux, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS13}, QuicConfig: quicConfig}
	go func() {
		if err := s.server.Serve(packet); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(common.NewError("http3 server failed").Base(err))
		}
	}()
	return s, nil
}
