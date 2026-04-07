package masque

import (
	"context"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

type OutboundConn struct {
	tunnel.Conn
	tcpConn net.Conn
}

func (c *OutboundConn) Read(p []byte) (int, error) {
	return c.Conn.Read(p)
}

func (c *OutboundConn) Write(p []byte) (int, error) {
	return c.Conn.Write(p)
}

func (c *OutboundConn) Close() error {
	return c.Conn.Close()
}

func (c *OutboundConn) LocalAddr() net.Addr {
	return c.tcpConn.LocalAddr()
}

func (c *OutboundConn) RemoteAddr() net.Addr {
	return c.tcpConn.RemoteAddr()
}

func (c *OutboundConn) SetDeadline(t time.Time) error {
	return c.tcpConn.SetDeadline(t)
}

func (c *OutboundConn) SetReadDeadline(t time.Time) error {
	return c.tcpConn.SetReadDeadline(t)
}

func (c *OutboundConn) SetWriteDeadline(t time.Time) error {
	return c.tcpConn.SetWriteDeadline(t)
}

func (c *OutboundConn) Metadata() *tunnel.Metadata {
	return nil
}

type InboundConn struct {
	OutboundConn
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *InboundConn) Close() error {
	c.cancel()
	return c.Conn.Close()
}
