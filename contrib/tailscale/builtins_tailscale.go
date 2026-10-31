//go:build b_tailscale
// +build b_tailscale

// Package tailscale exposes read-only access to the local Tailscale daemon.
//
// Build Rye with the b_tailscale tag (and b_contrib) to enable these builtins:
//
//	go build -tags "b_contrib,b_tailscale" .
package tailscale

import (
	"context"
	"net/netip"
	"sort"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"tailscale.com/client/local"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tailcfg"
	"tailscale.com/types/views"
)

var Builtins_tailscale = map[string]*env.Builtin{

	// Tests:
	// equal { tailscale/new |type? } 'native
	// equal { tailscale/new |kind? } 'tailscale-client
	// Returns:
	// * native tailscale-client connected to the default local daemon socket
	"new": {
		Argsn: 0,
		Doc:   "Creates a client for the local Tailscale daemon using the default socket.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, &local.Client{}, "tailscale-client")
		},
	},

	// Tests:
	// equal { tailscale/new-socket "/var/run/tailscale/tailscaled.sock" |kind? } 'tailscale-client
	// Args:
	// * socket: path to the tailscaled local API socket
	// Returns:
	// * native tailscale-client
	"new-socket": {
		Argsn: 1,
		Doc:   "Creates a client for the local Tailscale daemon using a specific socket path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch socket := arg0.(type) {
			case env.String:
				return *env.NewNative(ps.Idx, &local.Client{Socket: socket.Value}, "tailscale-client")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "new-socket")
			}
		},
	},

	// Tests:
	// ; requires a running local tailscaled
	// equal { tailscale/new |Status |type? } 'dict
	// Args:
	// * client: native tailscale-client
	// Returns:
	// * dict describing the current tailnet status, including a "peers" list
	"tailscale-client//Status": {
		Argsn: 1,
		Doc:   "Fetches the full status of the local tailnet, including peers, self, users and exit node.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, errObj := tailscaleClient(ps, arg0, "tailscale-client//Status")
			if errObj != nil {
				return errObj
			}
			status, err := client.Status(context.Background())
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Could not connect to local Tailscale daemon: "+err.Error(), "tailscale-client//Status")
			}
			return statusToDict(status)
		},
	},

	// Tests:
	// ; requires a running local tailscaled
	// equal { tailscale/new |Status-without-peers |type? } 'dict
	// Args:
	// * client: native tailscale-client
	// Returns:
	// * dict describing the current tailnet status, without peer details
	"tailscale-client//Status-without-peers": {
		Argsn: 1,
		Doc:   "Fetches tailnet status from the local daemon without peer information.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, errObj := tailscaleClient(ps, arg0, "tailscale-client//Status-without-peers")
			if errObj != nil {
				return errObj
			}
			status, err := client.StatusWithoutPeers(context.Background())
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Could not connect to local Tailscale daemon: "+err.Error(), "tailscale-client//Status-without-peers")
			}
			return statusToDict(status)
		},
	},

	// Tests:
	// ; requires a running local tailscaled
	// equal { tailscale/new |Peers |type? } 'block
	// Args:
	// * client: native tailscale-client
	// Returns:
	// * block of peer dicts
	"tailscale-client//Peers": {
		Argsn: 1,
		Doc:   "Returns only the peers visible to the local Tailscale daemon as a block of dicts.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, errObj := tailscaleClient(ps, arg0, "tailscale-client//Peers")
			if errObj != nil {
				return errObj
			}
			status, err := client.Status(context.Background())
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Could not connect to local Tailscale daemon: "+err.Error(), "tailscale-client//Peers")
			}
			return peersToBlock(status)
		},
	},

	// Tests:
	// ; requires a running local tailscaled
	// equal { tailscale/status |type? } 'dict
	// Returns:
	// * dict describing the current tailnet status
	"status": {
		Argsn: 0,
		Doc:   "Convenience function that fetches status from the default local Tailscale daemon.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			status, err := (&local.Client{}).Status(context.Background())
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Could not connect to local Tailscale daemon: "+err.Error(), "status")
			}
			return statusToDict(status)
		},
	},

	// Tests:
	// ; requires a running local tailscaled
	// equal { tailscale/peers |type? } 'block
	// Returns:
	// * block of peer dicts from the default local Tailscale daemon
	"peers": {
		Argsn: 0,
		Doc:   "Convenience function that returns peers from the default local Tailscale daemon.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			status, err := (&local.Client{}).Status(context.Background())
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Could not connect to local Tailscale daemon: "+err.Error(), "peers")
			}
			return peersToBlock(status)
		},
	},
}

func tailscaleClient(ps *env.ProgramState, obj env.Object, fn string) (*local.Client, env.Object) {
	native, ok := obj.(env.Native)
	if !ok {
		return nil, evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, fn)
	}
	client, ok := native.Value.(*local.Client)
	if !ok || client == nil {
		return nil, evaldo.MakeBuiltinError(ps, "Argument is not a tailscale-client.", fn)
	}
	return client, nil
}

func statusToDict(status *ipnstate.Status) *env.Dict {
	peers := peersToBlock(status)

	data := map[string]any{
		"version":          *env.NewString(status.Version),
		"tun":              *env.NewBoolean(status.TUN),
		"backend-state":    *env.NewString(status.BackendState),
		"have-node-key":    *env.NewBoolean(status.HaveNodeKey),
		"auth-url":         *env.NewString(status.AuthURL),
		"tailscale-ips":    addrList(status.TailscaleIPs),
		"health":           stringList(status.Health),
		"magic-dns-suffix": *env.NewString(status.MagicDNSSuffix),
		"cert-domains":     stringList(status.CertDomains),
		"peers":            peers,
		"peer":             peers, // alias matching the Go Status.Peer field name
		"users":            usersToBlock(status.User),
	}

	if status.Self != nil {
		data["self"] = *peerToDict(status.Self)
	} else {
		data["self"] = env.Void{}
	}

	if status.ExitNodeStatus != nil {
		data["exit-node-status"] = exitNodeStatusToDict(status.ExitNodeStatus)
	} else {
		data["exit-node-status"] = env.Void{}
	}

	if status.CurrentTailnet != nil {
		data["current-tailnet"] = *env.NewDict(map[string]any{
			"name":              *env.NewString(status.CurrentTailnet.Name),
			"magic-dns-suffix":  *env.NewString(status.CurrentTailnet.MagicDNSSuffix),
			"magic-dns-enabled": *env.NewBoolean(status.CurrentTailnet.MagicDNSEnabled),
		})
	} else {
		data["current-tailnet"] = env.Void{}
	}

	if status.ClientVersion != nil {
		data["client-version"] = *env.NewDict(map[string]any{
			"running-latest":         *env.NewBoolean(status.ClientVersion.RunningLatest),
			"latest-version":         *env.NewString(status.ClientVersion.LatestVersion),
			"urgent-security-update": *env.NewBoolean(status.ClientVersion.UrgentSecurityUpdate),
			"notify":                 *env.NewBoolean(status.ClientVersion.Notify),
			"notify-url":             *env.NewString(status.ClientVersion.NotifyURL),
			"notify-text":            *env.NewString(status.ClientVersion.NotifyText),
		})
	} else {
		data["client-version"] = env.Void{}
	}

	return env.NewDict(data)
}

func peersToBlock(status *ipnstate.Status) env.Block {
	keys := status.Peers()
	peers := make([]env.Object, 0, len(keys))
	for _, key := range keys {
		if peer := status.Peer[key]; peer != nil {
			peers = append(peers, *peerToDict(peer))
		}
	}
	return *env.NewBlock(*env.NewTSeries(peers))
}

func peerToDict(peer *ipnstate.PeerStatus) *env.Dict {
	data := map[string]any{
		"id":                 *env.NewString(string(peer.ID)),
		"public-key":         *env.NewString(peer.PublicKey.String()),
		"hostname":           *env.NewString(peer.HostName),
		"dns-name":           *env.NewString(peer.DNSName),
		"os":                 *env.NewString(peer.OS),
		"user-id":            *env.NewInteger(int64(peer.UserID)),
		"alt-sharer-user-id": *env.NewInteger(int64(peer.AltSharerUserID)),
		"tailscale-ips":      addrList(peer.TailscaleIPs),
		"allowed-ips":        prefixListFromView(peer.AllowedIPs),
		"tags":               stringListFromView(peer.Tags),
		"primary-routes":     prefixListFromView(peer.PrimaryRoutes),
		"addrs":              stringList(peer.Addrs),
		"cur-addr":           *env.NewString(peer.CurAddr),
		"relay":              *env.NewString(peer.Relay),
		"peer-relay":         *env.NewString(peer.PeerRelay),
		"rx-bytes":           *env.NewInteger(peer.RxBytes),
		"tx-bytes":           *env.NewInteger(peer.TxBytes),
		"created":            *env.NewString(formatTime(peer.Created)),
		"last-write":         *env.NewString(formatTime(peer.LastWrite)),
		"last-seen":          *env.NewString(formatTime(peer.LastSeen)),
		"last-handshake":     *env.NewString(formatTime(peer.LastHandshake)),
		"online":             *env.NewBoolean(peer.Online),
		"exit-node":          *env.NewBoolean(peer.ExitNode),
		"exit-node-option":   *env.NewBoolean(peer.ExitNodeOption),
		"active":             *env.NewBoolean(peer.Active),
		"peer-api-url":       stringList(peer.PeerAPIURL),
		"expired":            *env.NewBoolean(peer.Expired),
	}

	if peer.KeyExpiry != nil {
		data["key-expiry"] = *env.NewString(formatTime(*peer.KeyExpiry))
	} else {
		data["key-expiry"] = env.Void{}
	}

	if peer.Location != nil {
		data["location"] = *env.NewDict(map[string]any{
			"country":      *env.NewString(peer.Location.Country),
			"country-code": *env.NewString(peer.Location.CountryCode),
			"city":         *env.NewString(peer.Location.City),
			"city-code":    *env.NewString(peer.Location.CityCode),
			"latitude":     *env.NewDecimal(peer.Location.Latitude),
			"longitude":    *env.NewDecimal(peer.Location.Longitude),
			"priority":     *env.NewInteger(int64(peer.Location.Priority)),
		})
	} else {
		data["location"] = env.Void{}
	}

	return env.NewDict(data)
}

func usersToBlock(users map[tailcfg.UserID]tailcfg.UserProfile) env.Block {
	ids := make([]tailcfg.UserID, 0, len(users))
	for id := range users {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	result := make([]env.Object, 0, len(ids))
	for _, id := range ids {
		user := users[id]
		result = append(result, *env.NewDict(map[string]any{
			"id":              *env.NewInteger(int64(user.ID)),
			"login-name":      *env.NewString(user.LoginName),
			"display-name":    *env.NewString(user.DisplayName),
			"profile-pic-url": *env.NewString(user.ProfilePicURL),
		}))
	}
	return *env.NewBlock(*env.NewTSeries(result))
}

func exitNodeStatusToDict(status *ipnstate.ExitNodeStatus) *env.Dict {
	return env.NewDict(map[string]any{
		"id":            *env.NewString(string(status.ID)),
		"online":        *env.NewBoolean(status.Online),
		"tailscale-ips": prefixList(status.TailscaleIPs),
	})
}

func stringList(values []string) env.List {
	data := make([]any, len(values))
	for i, value := range values {
		data[i] = *env.NewString(value)
	}
	return *env.NewList(data)
}

func stringListFromView(values *views.Slice[string]) env.List {
	if values == nil {
		return *env.NewList([]any{})
	}
	return stringList(values.AsSlice())
}

func addrList(values []netip.Addr) env.List {
	data := make([]any, len(values))
	for i, value := range values {
		data[i] = *env.NewString(value.String())
	}
	return *env.NewList(data)
}

func prefixList(values []netip.Prefix) env.List {
	data := make([]any, len(values))
	for i, value := range values {
		data[i] = *env.NewString(value.String())
	}
	return *env.NewList(data)
}

func prefixListFromView(values *views.Slice[netip.Prefix]) env.List {
	if values == nil {
		return *env.NewList([]any{})
	}
	return prefixList(values.AsSlice())
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
