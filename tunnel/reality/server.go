package reality

import (
	"context"
	"encoding/hex"
	"net"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/redirector"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/websocket"

	reality "github.com/xtls/reality"
)

// Server is a REALITY server
type Server struct {
	fallbackAddress *tunnel.Address
	httpResp        []byte
	connChan        chan tunnel.Conn
	redir           *redirector.Redirector
	ctx             context.Context
	cancel          context.CancelFunc
	underlay        tunnel.Server

	// REALITY-specific
	dest        string
	privateKey  []byte
	shortIDs    map[[8]byte]bool
	serverNames map[string]bool
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.underlay.AcceptConn(&Tunnel{})
		if err != nil {
			select {
			case <-s.ctx.Done():
			default:
				log.Fatal(common.NewError("reality transport accept error" + err.Error()))
			}
			return
		}
		go func(conn net.Conn) {
			realityConfig := &reality.Config{
				DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
					return net.Dial(network, address)
				},
				Type:        "tcp",
				Dest:        s.dest,
				ServerNames: s.serverNames,
				PrivateKey:  s.privateKey,
				ShortIds:    s.shortIDs,
			}

			handshakeRewindConn := common.NewRewindConn(conn)
			handshakeRewindConn.SetBufferSize(2048)

			realityConn, err := reality.Server(context.Background(), handshakeRewindConn, realityConfig)
			handshakeRewindConn.StopBuffering()

			if err != nil {
				if strings.Contains(err.Error(), "first record does not look like a TLS handshake") {
					handshakeRewindConn.Rewind()
					log.Error(common.NewError("failed to perform reality handshake with " + conn.RemoteAddr().String() + ", redirecting").Base(err))
					if s.fallbackAddress != nil {
						s.redir.Redirect(&redirector.Redirection{
							InboundConn: handshakeRewindConn,
							RedirectTo:  s.fallbackAddress,
						})
					} else if s.httpResp != nil {
						handshakeRewindConn.Write(s.httpResp)
						handshakeRewindConn.Close()
					} else {
						handshakeRewindConn.Close()
					}
				} else {
					log.Error(common.NewError("reality handshake failed").Base(err))
				}
				return
			}

			log.Info("reality connection from", conn.RemoteAddr())
			state := realityConn.ConnectionState()
			log.Trace("reality handshake", reality.CipherSuiteName(state.CipherSuite), state.DidResume, state.NegotiatedProtocol)

			s.connChan <- &Conn{
				Conn: realityConn,
			}
		}(conn)
	}
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	if _, ok := overlay.(*websocket.Tunnel); ok {
		log.Warn("websocket over reality is experimental")
	}
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("reality server closed")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

// NewServer creates a REALITY server
func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)

	if cfg.Reality.Dest == "" {
		return nil, common.NewError("reality dest is required")
	}

	privateKey, err := hex.DecodeString(cfg.Reality.PrivateKey)
	if err != nil {
		return nil, common.NewError("invalid reality private key").Base(err)
	}

	shortIDs := make(map[[8]byte]bool)
	for _, idStr := range cfg.Reality.ShortIDs {
		idBytes, err := hex.DecodeString(idStr)
		if err != nil {
			return nil, common.NewError("invalid reality short id: " + idStr).Base(err)
		}
		var id [8]byte
		copy(id[:], idBytes)
		shortIDs[id] = true
	}

	serverNames := make(map[string]bool)
	for _, name := range cfg.Reality.ServerNames {
		serverNames[name] = true
	}

	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay:    underlay,
		dest:        cfg.Reality.Dest,
		privateKey:  privateKey,
		shortIDs:    shortIDs,
		serverNames: serverNames,
		connChan:    make(chan tunnel.Conn, 32),
		redir:       redirector.NewRedirector(ctx),
		ctx:         ctx,
		cancel:      cancel,
	}

	go server.acceptLoop()

	log.Debug("reality server created, dest:", cfg.Reality.Dest)
	return server, nil
}
