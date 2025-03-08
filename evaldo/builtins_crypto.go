//go:build !no_crypto
// +build !no_crypto

package evaldo

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/hex"
	"encoding/pem"
	"io"

	"crypto/x509"
	"fmt"
	"time"

	//"crypto/x509"
	// "x/crypto/pkcs12"

	"github.com/refaktor/rye/env"
	//"software.sslmate.com/src/go-pkcs12"
	"filippo.io/age"
	"golang.org/x/crypto/pkcs12"
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
		Doc:   "Decode string to bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				r, err := hex.DecodeString(addr.Value)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failure to decode string.", "string//to-bytes")
				}
				return *env.NewNative(ps.Idx, r, "bytes")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "string//to-bytes")
			}
		},
	},

	"bytes//to-string": {
		Argsn: 1,
		Doc:   "Encoding value to string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return env.NewString(hex.EncodeToString(addr.Value.([]byte)))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "bytes//to-string")
			}
		},
	},

	"Ed25519-pub-key//to-string": {
		Argsn: 1,
		Doc:   "Turns public key to string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return env.NewString(hex.EncodeToString(addr.Value.(ed25519.PublicKey)))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Ed25519-pub-key//to-string")
			}
		},
	},

	"Ed25519-priv-key//to-string": {
		Argsn: 1,
		Doc:   "Turns private key to string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return env.NewString(hex.EncodeToString(addr.Value.(ed25519.PrivateKey)))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Ed25519-priv-key//to-string")
			}
		},
	},

	"ed25519-generate-keys": {
		Argsn: 0,
		Doc:   "Generates private and public key, returns them in a block. Public first.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			keys := make([]env.Object, 2)
			puk, pvk, err := ed25519.GenerateKey(nil)
			if err != nil {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Failed to generate keys.", "ed25519-generate-keys")
			}
			keys[0] = *env.NewNative(ps.Idx, puk, "Ed25519-pub-key")
			keys[1] = *env.NewNative(ps.Idx, pvk, "Ed25519-priv-key")
			ser := *env.NewTSeries(keys)
			return *env.NewBlock(ser)
		},
	},

	"ed25519-private-key": {
		Argsn: 1,
		Doc:   "Creates private key from string or bytes.",
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
					return MakeBuiltinError(ps, "Error in decoding string.", "ed25519-private-key")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType, env.StringType}, "ed25519-private-key")
			}
			return *env.NewNative(ps.Idx, ed25519.PrivateKey(pkey), "Ed25519-priv-key")

		},
	},

	"ed25519-public-key": {
		Argsn: 1,
		Doc:   "Creates public key from string or bytes.",
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
					return MakeBuiltinError(ps, "Error in decoding string.", "ed25519-public-key")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType, env.StringType}, "ed25519-public-key")
			}
			return *env.NewNative(ps.Idx, ed25519.PublicKey(pkey), "Ed25519-pub-key")

		},
	},

	//    	sigb := ed25519.Sign(pvk, buffer)
	//    	pubb, _ := hex.DecodeString(pub)

	"Ed25519-priv-key//sign": {
		Argsn: 2,
		Doc:   "Signs string with private key. Returns signature in bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pvk := arg0.(type) {
			case env.Native:
				switch buff := arg1.(type) {
				case env.String:
					sigb := ed25519.Sign(pvk.Value.(ed25519.PrivateKey), []byte(buff.Value))
					return *env.NewNative(ps.Idx, sigb, "bytes")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.StringType}, "Ed25519-priv-key//sign")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Ed25519-priv-key//sign")
			}
		},
	},

	"sha512": {
		Argsn: 1,
		Doc:   "Calculates SHA512 on string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				h := sha512.New()
				h.Write([]byte(s.Value))
				bs := h.Sum(nil)
				return env.NewString(hex.EncodeToString(bs[:]))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "sha512")
			}
		},
	},

	//
	// ##### Age ##### "Age encryption/decryption and key generation"
	//
	// Tests:
	// equal { age-generate-keys |first |type? } 'native
	// equal { age-generate-keys |first |kind? } 'age-identity
	// equal { age-generate-keys |second |type? } 'native
	// equal { age-generate-keys |second |kind? } 'age-recipient
	"age-generate-keys": {
		Argsn: 0,
		Doc:   "Generates a new age key pair (identity and recipient).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			identity, err := age.GenerateX25519Identity()
			if err != nil {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Failed to generate key pair.", "age-generate-keys")
			}
			keys := make([]env.Object, 2)
			keys[0] = *env.NewNative(ps.Idx, identity, "age-identity")
			keys[1] = *env.NewNative(ps.Idx, identity.Recipient(), "age-recipient")
			ser := *env.NewTSeries(keys)
			return *env.NewBlock(ser)
		},
	},

	// Tests:
	// equal { age-identity "AGE-SECRET-KEY-1UMNMNLE5ADV4V0X8LRMG4GVWM3WJ7GVH6JP3J2XSRDFENLJVVX4SDLWXML" |type? } 'native
	// equal { age-identity "AGE-SECRET-KEY-1UMNMNLE5ADV4V0X8LRMG4GVWM3WJ7GVH6JP3J2XSRDFENLJVVX4SDLWXML" |kind? } 'age-identity
	// equal { age-identity "invalid" |disarm |type? } 'error
	"age-identity": {
		Argsn: 1,
		Doc:   "Creates an age identity from a string or bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var identity *age.X25519Identity
			var err error
			switch ident := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(ident.GetKind()) != "bytes" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "age-identity")
				}
				identity, err = age.ParseX25519Identity(hex.EncodeToString(ident.Value.([]byte)))
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error in parsing identity: "+err.Error(), "age-identity")
				}
			case env.String:
				identity, err = age.ParseX25519Identity(ident.Value)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error in decoding string: "+err.Error(), "age-identity")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType, env.StringType}, "age-identity")
			}
			return *env.NewNative(ps.Idx, identity, "age-identity")
		},
	},

	// Tests:
	// equal { age-recipient "age1zwya0qq8c824n5ncxppekrm4egk6gnvfhag6dmr87xjqaeuwlsgq68mqj4" |type? } 'native
	// equal { age-recipient "age1zwya0qq8c824n5ncxppekrm4egk6gnvfhag6dmr87xjqaeuwlsgq68mqj4" |kind? } 'age-recipient
	// equal { age-recipient "invalid" |disarm |type? } 'error
	"age-recipient": {
		Argsn: 1,
		Doc:   "Creates an age recipient from a string or bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var recipient *age.X25519Recipient
			var err error
			switch rec := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(rec.GetKind()) != "bytes" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "age-recipient")
				}
				recipient, err = age.ParseX25519Recipient(hex.EncodeToString(rec.Value.([]byte)))
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error in parsing recipient: "+err.Error(), "age-recipient")
				}
			case env.String:
				recipient, err = age.ParseX25519Recipient(rec.Value)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error in decoding string: "+err.Error(), "age-recipient")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType, env.StringType}, "age-recipient")
			}
			return *env.NewNative(ps.Idx, recipient, "age-recipient")
		},
	},

	// Tests:
	// equal {
	//     age-generate-keys |set! { identity recipient }
	//     "SUPER SECRET" |reader |age-encrypt recipient |age-decrypt identity |read\string
	// } "SUPER SECRET"
	// equal { "SUPER SECRET" |reader |age-encrypt "password" |age-decrypt "password" |read\string } "SUPER SECRET"
	"age-encrypt": {
		Argsn: 2,
		Doc:   "Encrypts a reader with age for the provided age recipient or string password and returns a reader.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(r.GetKind()) != "reader" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "age-encrypt")
				}
				reader := r.Value.(io.Reader)
				var recipient age.Recipient
				switch rec := arg1.(type) {
				case env.Native:
					if ps.Idx.GetWord(rec.GetKind()) != "age-recipient" {
						ps.FailureFlag = true
						return MakeArgError(ps, 1, []env.Type{env.NativeType}, "age-encrypt")
					}
					recipient = rec.Value.(*age.X25519Recipient)
				case env.String:
					var err error
					recipient, err = age.NewScryptRecipient(rec.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Error in creating recipient: "+err.Error(), "age-encrypt")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType, env.StringType}, "age-encrypt")
				}
				buf := new(bytes.Buffer)
				w, err := age.Encrypt(buf, recipient)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error in encrypting: "+err.Error(), "age-encrypt")
				}

				data, err := io.ReadAll(reader)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error reading from reader: "+err.Error(), "age-encrypt")
				}

				if _, err := w.Write(data); err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error in writing to buffer: "+err.Error(), "age-encrypt")
				}
				w.Close()
				return *env.NewNative(ps.Idx, bytes.NewReader(buf.Bytes()), "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "age-encrypt")
			}
		},
	},

	"age-decrypt": {
		Argsn: 2,
		Doc:   "Decrypts a reader with age with the provided age identity or string password and returns a reader.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(r.GetKind()) != "reader" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "age-decrypt")
				}
				reader := r.Value.(io.Reader)
				var identity age.Identity
				switch ident := arg1.(type) {
				case env.Native:
					if ps.Idx.GetWord(ident.GetKind()) != "age-identity" {
						ps.FailureFlag = true
						return MakeArgError(ps, 1, []env.Type{env.NativeType}, "age-decrypt")
					}
					identity = ident.Value.(*age.X25519Identity)
				case env.String:
					var err error
					identity, err = age.NewScryptIdentity(ident.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Error in creating identity: "+err.Error(), "age-decrypt")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType, env.StringType}, "age-decrypt")
				}
				decrypted, err := age.Decrypt(reader, identity)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error in decrypting: "+err.Error(), "age-decrypt")
				}
				return *env.NewNative(ps.Idx, decrypted, "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "age-decrypt")
			}
		},
	},

	// pkcs12

	"pkcs12-to-pem": {
		Argsn: 2,
		Doc:   "Converts a PKCS#12 (.p12) file bytes to PEM blocks using a password.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p12Data := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(p12Data.GetKind()) != "bytes" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pkcs12-to-pem")
				}
				switch password := arg1.(type) {
				case env.String:
					blocks, err := pkcs12.ToPEM(p12Data.Value.([]byte), password.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to convert .p12 to PEM: %v", err), "pkcs12-to-pem")
					}
					// Return a block of PEM blocks as Native objects
					objects := make([]env.Object, len(blocks))
					for i, block := range blocks {
						objects[i] = *env.NewNative(ps.Idx, block, "pem-block")
					}
					ser := *env.NewTSeries(objects)
					return *env.NewBlock(ser)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "pkcs12-to-pem")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pkcs12-to-pem")
			}
		},
	},

	/* "pkcs12-encode-to-memory": {
		Argsn: 3,
		Doc:   "Encodes a certificate and private key into a PKCS#12 (.p12) file in memory with a password.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pkcs12-encode-to-memory")
				}
				switch key := arg1.(type) {
				case env.Native:
					// Assume key is a private key (interface{} for simplicity; refine based on your needs)
					switch password := arg2.(type) {
					case env.String:
						p12Data, err := pkcs12.EncodeToMemory(cert.Value.(*x509.Certificate), key.Value, password.Value)
						if err != nil {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, fmt.Sprintf("Failed to encode to .p12: %v", err), "pkcs12-encode-to-memory")
						}
						return *env.NewNative(ps.Idx, p12Data, "Go-bytes")
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "pkcs12-encode-to-memory")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "pkcs12-encode-to-memory")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pkcs12-encode-to-memory")
			}
		},
	},*/

	"pkcs12-decode": {
		Argsn: 2,
		Doc:   "Decodes a PKCS#12 (.p12) file bytes into certificate and private key using a password.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p12Data := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(p12Data.GetKind()) != "bytes" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pkcs12-decode")
				}
				switch password := arg1.(type) {
				case env.String:
					key, cert, err := pkcs12.Decode(p12Data.Value.([]byte), password.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to decode .p12: %v", err), "pkcs12-decode")
					}
					// Return a block with [private key, certificate]
					objects := make([]env.Object, 2)
					objects[0] = *env.NewNative(ps.Idx, key, "private-key") // Generic private key type
					objects[1] = *env.NewNative(ps.Idx, cert, "x509-certificate")
					ser := *env.NewTSeries(objects)
					return *env.NewBlock(ser)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "pkcs12-decode")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pkcs12-decode")
			}
		},
	},

	"pem-block//block-type?": {
		Argsn: 1,
		Doc:   "Parses bytes into an X.509 certificate.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch certData := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(certData.GetKind()) != "pem-block" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-parse-certificate")
				}
				pb, ok := certData.Value.(*pem.Block)
				if ok {
					return *env.NewString(pb.Type)
				}
				return MakeBuiltinError(ps, fmt.Sprintf("Failed to parse certificate"), "x509-parse-certificate")

			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-parse-certificate")
			}
		},
	},

	"pem-block//headers?": {
		Argsn: 1,
		Doc:   "Parses bytes into an X.509 certificate.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch certData := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(certData.GetKind()) != "pem-block" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-parse-certificate")
				}
				pb, ok := certData.Value.(*pem.Block)
				if ok {
					headers := make(map[string]interface{}, len(pb.Headers))
					for k, v := range pb.Headers {
						headers[k] = v // string is automatically converted to interface{}
					}
					return *env.NewDict(headers)
				}
				return MakeBuiltinError(ps, fmt.Sprintf("Failed to parse certificate"), "x509-parse-certificate")

			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-parse-certificate")
			}
		},
	},

	"x509-parse-certificate": {
		Argsn: 1,
		Doc:   "Parses bytes into an X.509 certificate.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch certData := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(certData.GetKind()) != "pem-block" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-parse-certificate")
				}
				pb, ok := certData.Value.(*pem.Block)
				if ok {
					cert, err := x509.ParseCertificate(pb.Bytes)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to parse certificate: %v", err), "x509-parse-certificate")
					}
					return *env.NewNative(ps.Idx, cert, "x509-certificate")
				}
				return MakeBuiltinError(ps, fmt.Sprintf("Failed to parse certificate"), "x509-parse-certificate")

			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-parse-certificate")
			}
		},
	},

	"x509-certificate//not-after?": {
		Argsn: 1,
		Doc:   "Checks if an X.509 certificate has expired.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
				}
				c := cert.Value.(*x509.Certificate)
				return env.NewTime(c.NotAfter)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
			}
		},
	},

	"x509-certificate//not-before?": {
		Argsn: 1,
		Doc:   "Checks if an X.509 certificate has expired.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
				}
				c := cert.Value.(*x509.Certificate)
				return env.NewTime(c.NotBefore)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
			}
		},
	},

	"x509-certificate//is-expired": {
		Argsn: 1,
		Doc:   "Checks if an X.509 certificate has expired.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
				}
				c := cert.Value.(*x509.Certificate)
				if time.Now().After(c.NotAfter) {
					return env.Integer{1} // True
				}
				return env.Integer{0} // False
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
			}
		},
	},
}
