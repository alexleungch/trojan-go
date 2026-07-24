package quic

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	h3 "github.com/quic-go/quic-go/http3"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Client struct {
	transport  *h3.RoundTripper
	url        string
	enable0RTT bool
	remoteAddr net.Addr
}

func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	reader, writer := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, reader)
	if err != nil {
		cancel()
		return nil, common.NewError("http3 failed to create request").Base(err)
	}
	if c.enable0RTT {
		// 0-RTT data can be replayed. It is opt-in for deployments that accept that risk.
		req.Method = h3.MethodGet0RTT
	}
	req.Header.Set("Accept-Encoding", "identity")
	result := make(chan *http.Response, 1)
	errs := make(chan error, 1)
	go func() {
		resp, err := c.transport.RoundTrip(req)
		if err != nil {
			errs <- err
			return
		}
		result <- resp
	}()
	select {
	case resp := <-result:
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			writer.Close()
			cancel()
			return nil, common.NewError(fmt.Sprintf("http3 server returned status %d", resp.StatusCode))
		}
		return &Conn{reader: resp.Body, writer: writer, remoteAddr: c.remoteAddr}, nil
	case err := <-errs:
		writer.Close()
		cancel()
		return nil, common.NewError("http3 failed to open stream").Base(err)
	case <-ctx.Done():
		writer.Close()
		return nil, common.NewError("http3 connection cancelled").Base(ctx.Err())
	}
}
func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	return nil, common.NewError("http3 does not support packet tunnels")
}
func (c *Client) Close() error { return c.transport.Close() }

func NewClient(ctx context.Context, _ tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	if !cfg.HTTP3.Enabled {
		return nil, common.NewError("http3 is not enabled")
	}
	serverName := cfg.HTTP3.SNI
	if serverName == "" {
		serverName = cfg.RemoteHost
	}
	path := cfg.HTTP3.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		return nil, common.NewError("http3 path must start with /")
	}
	addr := tunnel.NewAddressFromHostPort("udp", cfg.RemoteHost, cfg.RemotePort)
	return &Client{transport: &h3.RoundTripper{TLSClientConfig: &tls.Config{ServerName: serverName, NextProtos: []string{h3.NextProtoH3}, MinVersion: tls.VersionTLS13}, DisableCompression: true}, url: "https://" + addr.String() + path, enable0RTT: cfg.HTTP3.Enable0RTT, remoteAddr: addr}, nil
}
