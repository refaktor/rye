//go:build b_maxmind
// +build b_maxmind

package maxmind

import (
    "net"

    "github.com/oschwald/geoip2-golang"
    "github.com/refaktor/rye/env"
    "github.com/refaktor/rye/evaldo"
)

// We expose simple open/close/lookup-country/lookup-asn builtins.
// Handles are env.Native with type tags: "maxmind-db"

type dbHandle struct{
    reader *geoip2.Reader
}

var Builtins_maxmind = map[string]*env.Builtin{
    // Generic URI Open: Open maxmind://path or Open maxmind//relative
    "maxmind-uri//Open": {
        Argsn: 1,
        Doc:   "Open a MaxMind .mmdb via URI (maxmind://). Returns native maxmind-db handle.",
        Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
            switch u := arg0.(type) {
            case env.Uri:
                r, err := geoip2.Open(u.GetPath())
                if err != nil {
                    return evaldo.MakeError(ps, err.Error())
                }
                return *env.NewNative(ps.Idx, &dbHandle{reader: r}, "maxmind-db")
            default:
                return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "maxmind-uri//Open")
            }
        },
    },

    "maxmind-db//Close": {
        Argsn: 1,
        Doc:   "Close a MaxMind DB handle",
        Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
            switch h := arg0.(type) {
            case env.Native:
                if dh, ok := h.Value.(*dbHandle); ok && dh.reader != nil {
                    dh.reader.Close()
                    dh.reader = nil
                    return env.Integer{1}
                }
                return evaldo.MakeError(ps, "Arg 1 not maxmind-db handle.")
            default:
                return evaldo.MakeError(ps, "Arg 1 not native.")
            }
        },
    },
    "maxmind-db//Lookup-country": {
        Argsn: 2,
        Doc:   "Lookup country for IP. Args: maxmind-db handle, String IP. Returns dict { ip country_code country_name }",
        Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
            switch h := arg0.(type) {
            case env.Native:
                dh, ok := h.Value.(*dbHandle)
                if !ok || dh.reader == nil {
                    return evaldo.MakeError(ps, "Arg 1 not maxmind-db handle.")
                }
                switch ipS := arg1.(type) {
                case env.String:
                    ip := net.ParseIP(ipS.Value)
                    if ip == nil {
                        return evaldo.MakeError(ps, "Invalid IP address.")
                    }
                    rec, err := dh.reader.Country(ip)
                    if err != nil {
                        return evaldo.MakeError(ps, err.Error())
                    }
                    data := map[string]any{
                        "ip":           ipS.Value,
                        "country_code": rec.Country.IsoCode,
                        "country_name": rec.Country.Names["en"],
                    }
                    return *env.NewDict(data)
                default:
                    return evaldo.MakeError(ps, "Arg 2 not String.")
                }
            default:
                return evaldo.MakeError(ps, "Arg 1 not native.")
            }
        },
    },
    "maxmind-db//Lookup-asn": {
        Argsn: 2,
        Doc:   "Lookup ASN for IP using GeoLite2-ASN DB. Args: maxmind-db handle, String IP. Returns dict { ip asn organization }",
        Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
            switch h := arg0.(type) {
            case env.Native:
                dh, ok := h.Value.(*dbHandle)
                if !ok || dh.reader == nil {
                    return evaldo.MakeError(ps, "Arg 1 not maxmind-db handle.")
                }
                switch ipS := arg1.(type) {
                case env.String:
                    ip := net.ParseIP(ipS.Value)
                    if ip == nil {
                        return evaldo.MakeError(ps, "Invalid IP address.")
                    }
                    rec, err := dh.reader.ASN(ip)
                    if err != nil {
                        return evaldo.MakeError(ps, err.Error())
                    }
                    data := map[string]any{
                        "ip":           ipS.Value,
                        "asn":          int64(rec.AutonomousSystemNumber),
                        "organization": rec.AutonomousSystemOrganization,
                    }
                    return *env.NewDict(data)
                default:
                    return evaldo.MakeError(ps, "Arg 2 not String.")
                }
            default:
                return evaldo.MakeError(ps, "Arg 1 not native.")
            }
        },
    },
}
