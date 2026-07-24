package quic

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

type writeCloser struct{ io.Writer }

func (writeCloser) Close() error { return nil }

// Conn presents one full-duplex HTTP/3 request/response pair as a tunnel connection.
type Conn struct {
	reader     io.ReadCloser
	writer     io.WriteCloser
	localAddr  net.Addr
	remoteAddr net.Addr
	closeOnce  sync.Once
	closeErr   error
}

func (c *Conn) Read(p []byte) (int, error)       { return c.reader.Read(p) }
func (c *Conn) Write(p []byte) (int, error)      { return c.writer.Write(p) }
func (c *Conn) LocalAddr() net.Addr              { return c.localAddr }
func (c *Conn) RemoteAddr() net.Addr             { return c.remoteAddr }
func (c *Conn) SetDeadline(time.Time) error      { return nil }
func (c *Conn) SetReadDeadline(time.Time) error  { return nil }
func (c *Conn) SetWriteDeadline(time.Time) error { return nil }
func (c *Conn) Metadata() *tunnel.Metadata       { return nil }
func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		c.closeErr = c.writer.Close()
		if err := c.reader.Close(); c.closeErr == nil {
			c.closeErr = err
		}
	})
	return c.closeErr
}
