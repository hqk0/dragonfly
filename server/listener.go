package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/df-mc/dragonfly/server/internal/legacyprotocol"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// Listener is a source for connections that may be listened on by a Server using Server.listen. Proxies can use this to
// provide players from a different source.
type Listener interface {
	// Accept blocks until the next connection is established and returns it. An error is returned if the Listener was
	// closed using Close.
	Accept() (session.Conn, error)
	// Disconnect disconnects a connection from the Listener with a reason.
	Disconnect(conn session.Conn, reason string) error
	io.Closer
}

// listenerFunc may be used to return a *minecraft.Listener using a Config. It
// is the standard listener used when UserConfig.Config() is called.
func (uc UserConfig) listenerFunc(conf Config) (Listener, error) {
	cfg := minecraft.ListenConfig{
		MaximumPlayers:         conf.MaxPlayers,
		StatusProvider:         conf.StatusProvider,
		AuthenticationDisabled: conf.AuthDisabled,
		ResourcePacks:          conf.Resources,
		TexturePacksRequired:   conf.ResourcesRequired,
		Compression:            conf.Compression,
		Allow:                  allowSupportedVersion(conf.Allower),
		AcceptedProtocols:      []minecraft.Protocol{legacyprotocol.Protocol{}},
	}
	if conf.Log.Enabled(context.Background(), slog.LevelDebug) {
		cfg.ErrorLog = conf.Log.With("net origin", "gophertunnel")
	}
	l, err := cfg.Listen("raknet", uc.Network.Address)
	if err != nil {
		return nil, fmt.Errorf("create minecraft listener: %w", err)
	}
	conf.Log.Info("Listener running.", "addr", l.Addr())
	return listener{l}, nil
}

// allowSupportedVersion rejects Minecraft 1.26.44 before post-login packets
// are written using the 1.26.40-1.26.43 protocol adapter. Minecraft 1.26.44
// shares protocol ID 2168 with those releases, so the game version in the
// login data is the only way to distinguish it.
func allowSupportedVersion(next Allower) func(net.Addr, login.IdentityData, login.ClientData) (string, bool) {
	return func(addr net.Addr, identityData login.IdentityData, clientData login.ClientData) (string, bool) {
		if clientData.GameVersion == "1.26.44" {
			return "Minecraft 1.26.44 is not supported. Please update to 1.26.45.", false
		}
		return next.Allow(addr, identityData, clientData)
	}
}

// listener is a Listener implementation that wraps around a minecraft.Listener so that it can be listened on by
// Server.
type listener struct {
	*minecraft.Listener
}

// Accept blocks until the next connection is established and returns it. An error is returned if the Listener was
// closed using Close.
func (l listener) Accept() (session.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return conn.(session.Conn), err
}

// Disconnect disconnects a connection from the Listener with a reason.
func (l listener) Disconnect(conn session.Conn, reason string) error {
	return l.Listener.Disconnect(conn.(*minecraft.Conn), reason)
}
