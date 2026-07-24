package transport

import (
	"context"
	"os"
	"os/exec"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/freedom"
)

// Client implements tunnel.Client
type Client struct {
	serverAddress *tunnel.Address
	cmd           *exec.Cmd
	ctx           context.Context
	cancel        context.CancelFunc
	direct        *freedom.Client
}

func (c *Client) Close() error {
	c.cancel()
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	return nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

// DialConn implements tunnel.Client. It will ignore the params and directly dial to the remote server
func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.direct.DialConn(c.serverAddress, nil)
	if err != nil {
		return nil, common.NewError("transport failed to connect to remote server").Base(err)
	}
	return &Conn{
		Conn: conn,
	}, nil
}

// NewClient creates a transport layer client
func NewClient(ctx context.Context, _ tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)

	var cmd *exec.Cmd
	serverAddress := tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort)

	if cfg.TransportPlugin.Enabled {
		log.Warn("trojan-go will use transport plugin and work in plain text mode")
		switch cfg.TransportPlugin.Type {
		case "other":
			cmd = exec.Command(cfg.TransportPlugin.Command, cfg.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, cfg.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			cmd.Start()
		case "plaintext":
			// do nothing
		default:
			return nil, common.NewError("invalid plugin type: " + cfg.TransportPlugin.Type)
		}
	}

	direct, err := freedom.NewClient(ctx, nil)
	common.Must(err)
	ctx, cancel := context.WithCancel(ctx)
	client := &Client{
		serverAddress: serverAddress,
		cmd:           cmd,
		ctx:           ctx,
		cancel:        cancel,
		direct:        direct,
	}
	return client, nil
}
