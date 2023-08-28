//go:build b_crypto
// +build b_crypto

package evaldo

import (
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/hex"
	"rye/env"
)

/* Our strategy to only support signed files

A) .codepks file holds public keys of developers we can trust, owner must be root and write access for all except root must be forbidden

B) .pubkeys are compiled into the rye binary. Only root can change it since it's in the /usr/bin, monitoring programs can check that no other rye runs

We make list of public keys and a flag "careful" as a global variable.

if a "careful" flag is set when a file is "done" the last five lines are checked for comment ;#codesig 123131321231 The fike up to this line is checked for signature

/*

    	priv := "e06d3183d14159228433ed599221b80bd0a5ce8352e4bdf0262f76786ef1c74db7e7a9fea2c0eb269d61e3b38e450a22e754941ac78479d6c54e1faf6037881d"
    	pub := "77ff84905a91936367c01360803104f92432fcd904a43511876df5cdf3e7e548"
    	sig := "6834284b6b24c3204eb2fea824d82f88883a3d95e8b4a21b8c0ded553d17d17ddf9a8a7104b1258f30bed3787e6cb896fca78c58f8e03b5f18f14951a87d9a08"
    	// d := hex.EncodeToString([]byte(priv))
    	privb, _ := hex.DecodeString(priv)
    	pvk := ed25519.PrivateKey(privb)
    	buffer := []byte("4:salt6:foobar3:seqi1e1:v12:Hello World!")
    	sigb := ed25519.Sign(pvk, buffer)
    	pubb, _ := hex.DecodeString(pub)
    	sigb2, _ := hex.DecodeString(sig)
    	log.Println(ed25519.Verify(pubb, buffer, sigb))
    	log.Printf("%x\n", pvk.Public())
    	log.Printf("%x\n", sigb)
    	log.Printf("%x\n", sigb2)

	priv: "e06d3183d14159228433ed599221b80bd0a5ce8352e4bdf0262f76786ef1c74db7e7a9fea2c0eb269d61e3b38e450a22e754941ac78479d6c54e1faf6037881d"
   	pub: "77ff84905a91936367c01360803104f92432fcd904a43511876df5cdf3e7e548"
   	sig: "6834284b6b24c3204eb2fea824d82f88883a3d95e8b4a21b8c0ded553d17d17ddf9a8a7104b1258f30bed3787e6cb896fca78c58f8e03b5f18f14951a87d9a08"

<rye-ed25519-privk> new-private-key priv
sign buffer pk
verify buffer pubk sigb
*/

var Builtins_crypto = map[string]*env.Builtin{

	"string//to-bytes": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				r, err := hex.DecodeString(addr.Value)
				if err != nil {
					env1.FailureFlag = true
					return env.NewError("failure to decode string")
				}
				return *env.NewNative(env1.Idx, r, "Go-bytes")
			default:
				env1.FailureFlag = true
				return env.NewError("arg 0 should be String")
			}
		},
	},

	"Go-bytes//to-string": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return env.String{hex.EncodeToString(addr.Value.([]byte))}
			default:
				env1.FailureFlag = true
				return env.NewError("arg 0 should be Native")
			}
		},
	},

	"Ed25519-pub-key//to-string": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return env.String{hex.EncodeToString(addr.Value.(ed25519.PublicKey))}
			default:
				env1.FailureFlag = true
				return env.NewError("arg 0 should be Native")
			}
		},
	},

	"Ed25519-priv-key//to-string": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return env.String{hex.EncodeToString(addr.Value.(ed25519.PrivateKey))}
			default:
				env1.FailureFlag = true
				return env.NewError("arg 0 should be Native")
			}
		},
	},

	"ed25519-generate-keys": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			keys := make([]env.Object, 2)
			puk, pvk, err := ed25519.GenerateKey(nil)
			if err != nil {
				ps.FailureFlag = true
				return env.NewError("failed to generate keys")
			}
			keys[0] = *env.NewNative(ps.Idx, ed25519.PublicKey(puk), "Ed25519-pub-key")
			keys[1] = *env.NewNative(ps.Idx, ed25519.PrivateKey(pvk), "Ed25519-priv-key")
			ser := *env.NewTSeries(keys)
			return *env.NewBlock(ser)
		},
	},

	"ed25519-private-key": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var pkey []byte
			var err error
			switch server := arg0.(type) {
			case env.Native:
				pkey = server.Value.([]byte)
			case env.String:
				pkey, err = hex.DecodeString(server.Value)
				if err != nil {
					ps.FailureFlag = true
					return env.NewError("decode err")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 should be string or native")
			}
			return *env.NewNative(ps.Idx, ed25519.PrivateKey(pkey), "Ed25519-priv-key")

		},
	},

	"ed25519-public-key": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var pkey []byte
			var err error
			switch server := arg0.(type) {
			case env.Native:
				pkey = server.Value.([]byte)
			case env.String:
				pkey, err = hex.DecodeString(server.Value)
				if err != nil {
					ps.FailureFlag = true
					return env.NewError("decode err")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 should be string or native")
			}
			return *env.NewNative(ps.Idx, ed25519.PublicKey(pkey), "Ed25519-pub-key")

		},
	},

	//    	sigb := ed25519.Sign(pvk, buffer)
	//    	pubb, _ := hex.DecodeString(pub)

	"Ed25519-priv-key//sign": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pvk := arg0.(type) {
			case env.Native:
				switch buff := arg1.(type) {
				case env.String:
					sigb := ed25519.Sign(pvk.Value.(ed25519.PrivateKey), []byte(buff.Value))
					return *env.NewNative(ps.Idx, sigb, "Go-bytes")
				default:
					ps.FailureFlag = true
					return env.NewError("arg 1 should be string")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 should be native")
			}
		},
	},

	"sha512": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				h := sha512.New()
				h.Write([]byte(s.Value))
				bs := h.Sum(nil)
				return env.String{hex.EncodeToString(bs[:])}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 1 should be String")
			}
		},
	},
}
