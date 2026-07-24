package server

import (
	"context"

	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/proxy/client"
	"github.com/p4gefau1t/trojan-go/tunnel/freedom"
	"github.com/p4gefau1t/trojan-go/tunnel/mux"
	"github.com/p4gefau1t/trojan-go/tunnel/quic"
	"github.com/p4gefau1t/trojan-go/tunnel/reality"
	"github.com/p4gefau1t/trojan-go/tunnel/router"
	"github.com/p4gefau1t/trojan-go/tunnel/simplesocks"
	"github.com/p4gefau1t/trojan-go/tunnel/tls"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
	"github.com/p4gefau1t/trojan-go/tunnel/websocket"
)

const Name = "SERVER"

func init() {
	proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
		cfg := config.FromContext(ctx, Name).(*client.Config)
		ctx, cancel := context.WithCancel(ctx)
		transportServer, err := transport.NewServer(ctx, nil)
		if err != nil {
			cancel()
			return nil, err
		}
		clientStack := []string{freedom.Name}
		if cfg.Router.Enabled {
			clientStack = []string{freedom.Name, router.Name}
		}

		root := &proxy.Node{
			Name:       transport.Name,
			Next:       make(map[string]*proxy.Node),
			IsEndpoint: false,
			Context:    ctx,
			Server:     transportServer,
		}

		if !cfg.TransportPlugin.Enabled {
			if cfg.Reality.Enabled {
				root = root.BuildNext(reality.Name)
			} else {
				root = root.BuildNext(tls.Name)
			}
		}

		root.BuildNext(trojan.Name).BuildNext(mux.Name).BuildNext(simplesocks.Name).IsEndpoint = true
		root.BuildNext(trojan.Name).IsEndpoint = true

		wsSubTree := root.BuildNext(websocket.Name)
		wsSubTree.BuildNext(trojan.Name).BuildNext(mux.Name).BuildNext(simplesocks.Name).IsEndpoint = true
		wsSubTree.BuildNext(trojan.Name).IsEndpoint = true

		serverList := proxy.FindAllEndpoints(root)
		if cfg.HTTP3.Enabled {
			http3Server, err := quic.NewServer(ctx, nil)
			if err != nil {
				cancel()
				return nil, err
			}
			http3Root := &proxy.Node{Name: quic.Name, Next: make(map[string]*proxy.Node), Context: ctx, Server: http3Server}
			http3Root.BuildNext(trojan.Name).BuildNext(mux.Name).BuildNext(simplesocks.Name).IsEndpoint = true
			http3Root.BuildNext(trojan.Name).IsEndpoint = true
			serverList = append(serverList, proxy.FindAllEndpoints(http3Root)...)
		}
		clientList, err := proxy.CreateClientStack(ctx, clientStack)
		if err != nil {
			cancel()
			return nil, err
		}
		return proxy.NewProxy(ctx, cancel, serverList, clientList), nil
	})
}
