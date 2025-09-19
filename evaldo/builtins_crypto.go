//go:build !no_crypto
// +build !no_crypto

package evaldo

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"io"
	"math/big"

	"crypto/x509"
	"fmt"
	"time"

	//"crypto/x509"
	// "x/crypto/pkcs12"

	"github.com/refaktor/rye/env"
	sslmate "software.sslmate.com/src/go-pkcs12"

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

	// Tests:
	//  equal { cc crypto "48656c6c6f20776f726c64" |decode\hex |type? } 'native
	//  equal { cc crypto "48656c6c6f20776f726c64" |decode\hex |kind? } 'bytes
	//  equal { cc crypto "invalid" |decode\hex |disarm |type? } 'error
	// Args:
	// * hex-string: hexadecimal encoded string to decode
	// Returns:
	// * native bytes object containing the decoded data
	"decode\\hex": {
		Argsn: 1,
		Doc:   "Decodes a hexadecimal string to a bytes native value.",
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

	// Tests:
	//  equal { cc crypto "48656c6c6f20776f726c64" |decode\hex |encode-to\hex } "48656c6c6f20776f726c64"
	//  equal { cc crypto "Hello world" |sha512 |decode\hex |encode-to\hex |type? } 'string
	// Args:
	// * bytes: native bytes object to encode
	// Returns:
	// * string containing the hexadecimal representation of the bytes
	"encode-to\\hex": {
		Argsn: 1,
		Doc:   "Encodes a bytes native value to a hexadecimal string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return *env.NewString(hex.EncodeToString(addr.Value.([]byte)))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "bytes//to-string")
			}
		},
	},

	// Tests:
	//  equal { cc crypto ed25519-generate-keys |first |to-string |type? } 'string
	// Args:
	// * key: Ed25519 public key as a native value
	// Returns:
	// * string containing the hexadecimal representation of the public key
	"Ed25519-pub-key//To-string": {
		Argsn: 1,
		Doc:   "Converts an Ed25519 public key to its hexadecimal string representation.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return *env.NewString(hex.EncodeToString(addr.Value.(ed25519.PublicKey)))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Ed25519-pub-key//To-string")
			}
		},
	},

	// Tests:
	//  equal { cc crypto ed25519-generate-keys |second |to-string |type? } 'string
	// Args:
	// * key: Ed25519 private key as a native value
	// Returns:
	// * string containing the hexadecimal representation of the private key
	"Ed25519-priv-key//To-string": {
		Argsn: 1,
		Doc:   "Converts an Ed25519 private key to its hexadecimal string representation.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Native:
				return *env.NewString(hex.EncodeToString(addr.Value.(ed25519.PrivateKey)))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Ed25519-priv-key//To-string")
			}
		},
	},

	// Tests:
	//  equal { cc crypto ed25519-generate-keys |type? } 'block
	//  equal { cc crypto ed25519-generate-keys |length? } 2
	//  equal { cc crypto ed25519-generate-keys |first |type? } 'native
	//  equal { cc crypto ed25519-generate-keys |first |kind? } 'Ed25519-pub-key
	//  equal { cc crypto ed25519-generate-keys |second |type? } 'native
	//  equal { cc crypto ed25519-generate-keys |second |kind? } 'Ed25519-priv-key
	// Args:
	// * none
	// Returns:
	// * block containing [public-key, private-key] as native values
	"ed25519-generate-keys": {
		Argsn: 0,
		Doc:   "Generates a new Ed25519 key pair and returns them in a block with public key first, then private key.",
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

	// Tests:
	//  ; equal { cc crypto ed25519-generate-keys |second |ed25519-private-key |type? } 'native
	//  ; equal { cc crypto ed25519-generate-keys |second |to-string |probe |ed25519-private-key |kind? } 'Ed25519-priv-key
	//  equal { cc crypto "invalid" |ed25519-private-key |disarm |type? } 'error
	// Args:
	// * key-data: string containing hexadecimal representation of the key or bytes native value
	// Returns:
	// * Ed25519 private key as a native value
	"ed25519-private-key": {
		Argsn: 1,
		Doc:   "Creates an Ed25519 private key from a hexadecimal string or bytes value.",
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

	// Tests:
	//  ; equal { cc crypto ed25519-generate-keys |first |to-string |ed25519-public-key |type? } 'native
	//  ; equal { cc crypto ed25519-generate-keys |first |to-string |ed25519-public-key |kind? } 'Ed25519-pub-key
	//  equal { cc crypto "invalid" |ed25519-public-key |disarm |type? } 'error
	// Args:
	// * key-data: string containing hexadecimal representation of the key or bytes native value
	// Returns:
	// * Ed25519 public key as a native value
	"ed25519-public-key": {
		Argsn: 1,
		Doc:   "Creates an Ed25519 public key from a hexadecimal string or bytes value.",
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

	// Tests:
	//  ; equal { cc crypto
	//  ;   ed25519-generate-keys |set! { pub priv }
	//  ;   "Hello world" priv |sign |type?
	//  ; } 'native
	//  ; equal { cc crypto
	//  ;   ed25519-generate-keys |set! { pub priv }
	//  ;  "Hello world" priv |sign |kind?
	//  ; } 'bytes
	// Args:
	// * key: Ed25519 private key as a native value
	// * message: string to sign
	// Returns:
	// * signature as a native bytes value
	"Ed25519-priv-key//Sign": {
		Argsn: 2,
		Doc:   "Signs a string message with an Ed25519 private key and returns the signature as bytes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pvk := arg0.(type) {
			case env.Native:
				switch buff := arg1.(type) {
				case env.String:
					sigb := ed25519.Sign(pvk.Value.(ed25519.PrivateKey), []byte(buff.Value))
					return *env.NewNative(ps.Idx, sigb, "bytes")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.StringType}, "Ed25519-priv-key//Sign")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Ed25519-priv-key//Sign")
			}
		},
	},

	// Tests:
	//  equal { cc crypto "Hello world" |sha512 |type? } 'string
	//  equal { cc crypto "Hello world" |sha512 |length? } 128
	//  equal { cc crypto "" |sha512 } "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	// Args:
	// * input: string to hash
	// Returns:
	// * string containing the hexadecimal representation of the SHA-512 hash
	"sha512": {
		Argsn: 1,
		Doc:   "Calculates the SHA-512 hash of a string and returns the result as a hexadecimal string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				h := sha512.New()
				h.Write([]byte(s.Value))
				bs := h.Sum(nil)
				return *env.NewString(hex.EncodeToString(bs[:]))
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
	// equal { cc crypto age-generate-keys |first |type? } 'native
	// equal { cc crypto age-generate-keys |first |kind? } 'age-identity
	// equal { cc crypto age-generate-keys |second |type? } 'native
	// equal { cc crypto age-generate-keys |second |kind? } 'age-recipient
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
	// equal { cc crypto age-identity "AGE-SECRET-KEY-1UMNMNLE5ADV4V0X8LRMG4GVWM3WJ7GVH6JP3J2XSRDFENLJVVX4SDLWXML" |type? } 'native
	// equal { cc crypto age-identity "AGE-SECRET-KEY-1UMNMNLE5ADV4V0X8LRMG4GVWM3WJ7GVH6JP3J2XSRDFENLJVVX4SDLWXML" |kind? } 'age-identity
	// equal { cc crypto age-identity "invalid" |disarm |type? } 'error
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
	// equal { cc crypto age-recipient "age1zwya0qq8c824n5ncxppekrm4egk6gnvfhag6dmr87xjqaeuwlsgq68mqj4" |type? } 'native
	// equal { cc crypto age-recipient "age1zwya0qq8c824n5ncxppekrm4egk6gnvfhag6dmr87xjqaeuwlsgq68mqj4" |kind? } 'age-recipient
	// equal { cc crypto age-recipient "invalid" |disarm |type? } 'error
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
	// equal { cc crypto
	//     age-generate-keys |set! { identity recipient }
	//     "SUPER SECRET" |reader |age-encrypt recipient |age-decrypt identity |Read\string
	// } "SUPER SECRET"
	// equal { cc crypto "SUPER SECRET" |reader |age-encrypt "password" |age-decrypt "password" |Read\string } "SUPER SECRET"
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

	// Tests:
	// equal { cc crypto
	//     age-generate-keys |set! { identity recipient }
	//     "SUPER SECRET" |reader |age-encrypt recipient |age-decrypt identity |Read\string
	// } "SUPER SECRET"
	// equal { cc crypto "SUPER SECRET" |reader |age-encrypt "password" |age-decrypt "password" |Read\string } "SUPER SECRET"
	// Args:
	// * reader: encrypted data as a reader native value
	// * identity-or-password: age identity native value or password string
	// Returns:
	// * decrypted data as a reader native value
	"age-decrypt": {
		Argsn: 2,
		Doc:   "Decrypts a reader with age using the provided age identity or string password and returns a reader with the decrypted content.",
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

	// Tests:
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-to-pem |type? } 'block
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-to-pem |first |type? } 'native
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-to-pem |first |kind? } 'pem-block
	// Args:
	// * p12-data: PKCS#12 file content as bytes native value
	// * password: string password for the PKCS#12 file
	// Returns:
	// * block containing PEM blocks as native values
	"pkcs12-to-pem": {
		Argsn: 2,
		Doc:   "Converts a PKCS#12 (.p12) file bytes to PEM blocks using the provided password. Returns a block of pem-block native values.",
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

	// Tests:
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-decode |type? } 'block
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-decode |length? } 3
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-decode |first |type? } 'native
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-decode |first |kind? } 'private-key
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-decode |second |type? } 'block
	// ; equal { cc crypto "cert.p12" read-file "password" pkcs12-decode |third |type? } 'block
	// Args:
	// * p12-data: PKCS#12 file content as bytes native value
	// * password: string password for the PKCS#12 file
	// Returns:
	// * block containing [private-key, certificates-block, ca-certificates-block] as native values
	"pkcs12-decode": {
		Argsn: 2,
		Doc:   "Decodes a PKCS#12 (.p12) file bytes into private key and certificates using the provided password. Returns a block with [private-key, certificates, ca-certificates].",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p12Data := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(p12Data.GetKind()) != "bytes" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "pkcs12-decode")
				}
				switch password := arg1.(type) {
				case env.String:
					key, cert, caCerts, err := sslmate.DecodeChain(p12Data.Value.([]byte), password.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to decode .p12: %v", err), "pkcs12-decode")
					}

					// Create certificates block
					certObjects := make([]env.Object, 0)
					if cert != nil {
						certObjects = append(certObjects, *env.NewNative(ps.Idx, cert, "x509-certificate"))
					}
					certSer := *env.NewTSeries(certObjects)
					certBlock := *env.NewBlock(certSer)

					// Create CA certificates block
					caCertObjects := make([]env.Object, len(caCerts))
					for i, caCert := range caCerts {
						caCertObjects[i] = *env.NewNative(ps.Idx, caCert, "x509-certificate")
					}
					caCertSer := *env.NewTSeries(caCertObjects)
					caCertBlock := *env.NewBlock(caCertSer)

					// Return a block with [private key, certificates, ca-certificates]
					objects := make([]env.Object, 3)
					if key != nil {
						objects[0] = *env.NewNative(ps.Idx, key, "private-key")
					} else {
						objects[0] = *env.NewNative(ps.Idx, nil, "private-key")
					}
					objects[1] = certBlock
					objects[2] = caCertBlock
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

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |block-type? } "CERTIFICATE"
	// Args:
	// * pem-block: PEM block as a native value
	// Returns:
	// * string containing the block type (e.g., "CERTIFICATE", "RSA PRIVATE KEY")
	"pem-block//Block-type?": {
		Argsn: 1,
		Doc:   "Returns the type of a PEM block as a string (e.g., 'CERTIFICATE', 'RSA PRIVATE KEY').",
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

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |headers? |type? } 'dict
	// Args:
	// * pem-block: PEM block as a native value
	// Returns:
	// * dictionary containing the PEM block headers
	"pem-block//Headers?": {
		Argsn: 1,
		Doc:   "Returns the headers of a PEM block as a dictionary.",
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

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |type? } 'native
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |kind? } 'x509-certificate
	// Args:
	// * pem-block: PEM block as a native value containing a certificate
	// Returns:
	// * X.509 certificate as a native value
	"x509-parse-certificate": {
		Argsn: 1,
		Doc:   "Parses a PEM block into an X.509 certificate native value.",
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

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |not-after? |type? } 'time
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * time value representing the certificate's expiration date
	"x509-certificate//Not-after?": {
		Argsn: 1,
		Doc:   "Returns the expiration date (NotAfter) of an X.509 certificate as a time value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
				}
				c := cert.Value.(*x509.Certificate)
				return *env.NewTime(c.NotAfter)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |not-before? |type? } 'time
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * time value representing the certificate's start date
	"x509-certificate//Not-before?": {
		Argsn: 1,
		Doc:   "Returns the start date (NotBefore) of an X.509 certificate as a time value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
				}
				c := cert.Value.(*x509.Certificate)
				return *env.NewTime(c.NotBefore)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//is-expired")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |is-expired |type? } 'integer
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |is-expired } 0 ; assuming cert is not expired
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * integer 1 if the certificate has expired, 0 otherwise
	"x509-certificate//Is-expired": {
		Argsn: 1,
		Doc:   "Checks if an X.509 certificate has expired. Returns 1 if expired, 0 otherwise.",
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

	// Tests:
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |type? } 'block
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |length? } 2
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |at 0 |type? } 'native
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |at 0 |kind? } 'x509-certificate
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |at 1 |type? } 'native
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |at 1 |kind? } 'rsa-private-key
	// ; equal { cc crypto generate-self-signed-certificate 1024 { "CommonName" "test.com" } |disarm |type? } 'error
	// Args:
	// * key-size: integer, must be at least 2048 bits
	// * subject: dictionary with fields like "CommonName" and "Organization"
	// Returns:
	// * block containing [certificate, private-key] as native values
	"generate-self-signed-certificate": {
		Argsn: 2,
		Doc:   "Generates a self-signed X.509 certificate with a new RSA key pair.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Arg0: Key size (int)
			switch keySize := arg0.(type) {
			case env.Integer:
				if keySize.Value < 2048 {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Key size must be at least 2048 bits", "generate-self-signed-certificate")
				}
				// Arg1: Subject info (dict with fields like "CommonName", "Organization", etc.)
				switch subjectDict := arg1.(type) {
				case env.Dict:
					// Generate RSA key pair
					privateKey, err := rsa.GenerateKey(rand.Reader, int(keySize.Value))
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to generate RSA key: %v", err), "generate-self-signed-certificate")
					}

					// Set up certificate template
					serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128)) // Random 128-bit serial number
					template := x509.Certificate{
						SerialNumber: serialNumber,
						Subject:      pkix.Name{},
						NotBefore:    time.Now(),
						NotAfter:     time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
						KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
						ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
						IsCA:         true, // Self-signed, so it’s a CA
					}

					// Populate subject from dict
					if cn, ok := subjectDict.Data["CommonName"]; ok {
						if cnStr, ok := cn.(string); ok {
							template.Subject.CommonName = cnStr
						}
					}
					if org, ok := subjectDict.Data["Organization"]; ok {
						if orgStr, ok := org.(string); ok {
							template.Subject.Organization = []string{orgStr}
						}
					}
					// Add more fields (Country, Locality, etc.) as needed

					// Create self-signed certificate
					certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to create certificate: %v", err), "generate-self-signed-certificate")
					}

					// Return block with [certificate, private-key]
					objects := []env.Object{
						*env.NewNative(ps.Idx, &x509.Certificate{Raw: certBytes}, "x509-certificate"),
						*env.NewNative(ps.Idx, privateKey, "rsa-private-key"),
					}
					ser := *env.NewTSeries(objects)
					return *env.NewBlock(ser)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.DictType}, "generate-self-signed-certificate")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "generate-self-signed-certificate")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key } cert key encode-to-pem |type? } 'block
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key }   cert key encode-to-pem |length? } 2
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key }   cert key encode-to-pem |at 0 |type? } 'native
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key }   cert key encode-to-pem |at 0 |kind? } 'Go-bytes
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key }   cert key encode-to-pem |at 1 |type? } 'native
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key }   cert key encode-to-pem |at 1 |kind? } 'Go-bytes
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key }   cert key encode-to-pem |at 0 "data/cert.pem" write-file "data/cert.pem" read-file |x509-parse-certificate |is-expired } 0
	// Args:
	// * certificate: X.509 certificate as a native value
	// * private-key: RSA private key as a native value
	// Returns:
	// * block with [cert-bytes, key-bytes] as Go-bytes native values
	"encode-to-pem": {
		Argsn: 2,
		Doc:   "Encodes a certificate and private key as PEM-formatted data.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch certObj := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(certObj.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "encode-to-pem")
				}
				cert := certObj.Value.(*x509.Certificate)
				switch keyObj := arg1.(type) {
				case env.Native:
					if ps.Idx.GetWord(keyObj.GetKind()) != "rsa-private-key" {
						ps.FailureFlag = true
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "encode-to-pem")
					}
					privateKey := keyObj.Value.(*rsa.PrivateKey)
					// Encode certificate to PEM
					certPEM := &bytes.Buffer{}
					err := pem.Encode(certPEM, &pem.Block{
						Type:  "CERTIFICATE",
						Bytes: cert.Raw,
					})
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to encode certificate to PEM: %v", err), "encode-to-pem")
					}
					// Encode private key to PEM
					keyPEM := &bytes.Buffer{}
					err = pem.Encode(keyPEM, &pem.Block{
						Type:  "RSA PRIVATE KEY",
						Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
					})
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to encode private key to PEM: %v", err), "encode-to-pem")
					}
					// Return block with [cert-bytes, key-bytes]
					objects := []env.Object{
						*env.NewNative(ps.Idx, certPEM.Bytes(), "Go-bytes"),
						*env.NewNative(ps.Idx, keyPEM.Bytes(), "Go-bytes"),
					}
					ser := *env.NewTSeries(objects)
					return *env.NewBlock(ser)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "encode-to-pem")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "encode-to-pem")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key } cert key "password" encode-to-p12 |type? } 'native
	// ; equal { cc crypto generate-self-signed-certificate 2048 { "CommonName" "test.com" } |set! { cert key } cert key "password" encode-to-p12 |kind? } 'Go-bytes
	// Args:
	// * certificate: X.509 certificate as a native value
	// * private-key: RSA private key as a native value
	// * password: string password to protect the PKCS#12 file
	// Returns:
	// * PKCS#12 encoded data as Go-bytes native value
	"encode-to-p12": {
		Argsn: 3,
		Doc:   "Encodes a certificate and private key into a PKCS#12 (.p12) file with password protection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch certObj := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(certObj.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "encode-to-p12")
				}
				cert := certObj.Value.(*x509.Certificate)
				switch keyObj := arg1.(type) {
				case env.Native:
					if ps.Idx.GetWord(keyObj.GetKind()) != "rsa-private-key" {
						ps.FailureFlag = true
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "encode-to-p12")
					}
					privateKey := keyObj.Value.(*rsa.PrivateKey)
					switch password := arg2.(type) {
					case env.String:
						p12Bytes, err := sslmate.Encode(rand.Reader, privateKey, cert, nil, password.Value)
						if err != nil {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, fmt.Sprintf("Failed to encode to .p12: %v", err), "encode-to-p12")
						}
						return *env.NewNative(ps.Idx, p12Bytes, "Go-bytes")
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "encode-to-p12")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "encode-to-p12")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "encode-to-p12")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |subject? |type? } 'string
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * string containing the certificate's subject DN
	"x509-certificate//Subject?": {
		Argsn: 1,
		Doc:   "Returns the subject Distinguished Name (DN) of an X.509 certificate as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//subject?")
				}
				c := cert.Value.(*x509.Certificate)
				return *env.NewString(c.Subject.String())
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//subject?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |issuer? |type? } 'string
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * string containing the certificate's issuer DN
	"x509-certificate//Issuer?": {
		Argsn: 1,
		Doc:   "Returns the issuer Distinguished Name (DN) of an X.509 certificate as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//issuer?")
				}
				c := cert.Value.(*x509.Certificate)
				return *env.NewString(c.Issuer.String())
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//issuer?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |serial-number? |type? } 'string
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * string containing the certificate's serial number
	"x509-certificate//Serial-number?": {
		Argsn: 1,
		Doc:   "Returns the serial number of an X.509 certificate as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//serial-number?")
				}
				c := cert.Value.(*x509.Certificate)
				return *env.NewString(c.SerialNumber.String())
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//serial-number?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |signature-algorithm? |type? } 'string
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * string containing the certificate's signature algorithm
	"x509-certificate//Signature-algorithm?": {
		Argsn: 1,
		Doc:   "Returns the signature algorithm of an X.509 certificate as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//signature-algorithm?")
				}
				c := cert.Value.(*x509.Certificate)
				return *env.NewString(c.SignatureAlgorithm.String())
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//signature-algorithm?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |public-key-algorithm? |type? } 'string
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * string containing the certificate's public key algorithm
	"x509-certificate//Public-key-algorithm?": {
		Argsn: 1,
		Doc:   "Returns the public key algorithm of an X.509 certificate as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//public-key-algorithm?")
				}
				c := cert.Value.(*x509.Certificate)
				return *env.NewString(c.PublicKeyAlgorithm.String())
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//public-key-algorithm?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |key-usage? |type? } 'block
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * block containing strings of key usage flags
	"x509-certificate//Key-usage?": {
		Argsn: 1,
		Doc:   "Returns the key usage flags of an X.509 certificate as a block of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//key-usage?")
				}
				c := cert.Value.(*x509.Certificate)
				var usages []env.Object
				if c.KeyUsage&x509.KeyUsageDigitalSignature > 0 {
					usages = append(usages, *env.NewString("Digital Signature"))
				}
				if c.KeyUsage&x509.KeyUsageContentCommitment > 0 {
					usages = append(usages, *env.NewString("Content Commitment"))
				}
				if c.KeyUsage&x509.KeyUsageKeyEncipherment > 0 {
					usages = append(usages, *env.NewString("Key Encipherment"))
				}
				if c.KeyUsage&x509.KeyUsageDataEncipherment > 0 {
					usages = append(usages, *env.NewString("Data Encipherment"))
				}
				if c.KeyUsage&x509.KeyUsageKeyAgreement > 0 {
					usages = append(usages, *env.NewString("Key Agreement"))
				}
				if c.KeyUsage&x509.KeyUsageCertSign > 0 {
					usages = append(usages, *env.NewString("Cert Sign"))
				}
				if c.KeyUsage&x509.KeyUsageCRLSign > 0 {
					usages = append(usages, *env.NewString("CRL Sign"))
				}
				if c.KeyUsage&x509.KeyUsageEncipherOnly > 0 {
					usages = append(usages, *env.NewString("Encipher Only"))
				}
				if c.KeyUsage&x509.KeyUsageDecipherOnly > 0 {
					usages = append(usages, *env.NewString("Decipher Only"))
				}
				ser := *env.NewTSeries(usages)
				return *env.NewBlock(ser)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//key-usage?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |extended-key-usage? |type? } 'block
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * block containing strings of extended key usage flags
	"x509-certificate//Extended-key-usage?": {
		Argsn: 1,
		Doc:   "Returns the extended key usage flags of an X.509 certificate as a block of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//extended-key-usage?")
				}
				c := cert.Value.(*x509.Certificate)
				usages := make([]env.Object, len(c.ExtKeyUsage))
				for i, extKeyUsage := range c.ExtKeyUsage {
					var usageStr string
					switch extKeyUsage {
					case x509.ExtKeyUsageServerAuth:
						usageStr = "Server Authentication"
					case x509.ExtKeyUsageClientAuth:
						usageStr = "Client Authentication"
					case x509.ExtKeyUsageCodeSigning:
						usageStr = "Code Signing"
					case x509.ExtKeyUsageEmailProtection:
						usageStr = "Email Protection"
					case x509.ExtKeyUsageTimeStamping:
						usageStr = "Time Stamping"
					case x509.ExtKeyUsageOCSPSigning:
						usageStr = "OCSP Signing"
					default:
						usageStr = fmt.Sprintf("Unknown (%d)", int(extKeyUsage))
					}
					usages[i] = *env.NewString(usageStr)
				}
				ser := *env.NewTSeries(usages)
				return *env.NewBlock(ser)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//extended-key-usage?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |dns-names? |type? } 'block
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * block containing DNS names from Subject Alternative Names
	"x509-certificate//Dns-names?": {
		Argsn: 1,
		Doc:   "Returns the DNS names from Subject Alternative Names of an X.509 certificate as a block of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//dns-names?")
				}
				c := cert.Value.(*x509.Certificate)
				names := make([]env.Object, len(c.DNSNames))
				for i, name := range c.DNSNames {
					names[i] = *env.NewString(name)
				}
				ser := *env.NewTSeries(names)
				return *env.NewBlock(ser)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//dns-names?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |ip-addresses? |type? } 'block
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * block containing IP addresses from Subject Alternative Names
	"x509-certificate//Ip-addresses?": {
		Argsn: 1,
		Doc:   "Returns the IP addresses from Subject Alternative Names of an X.509 certificate as a block of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//ip-addresses?")
				}
				c := cert.Value.(*x509.Certificate)
				addresses := make([]env.Object, len(c.IPAddresses))
				for i, addr := range c.IPAddresses {
					addresses[i] = *env.NewString(addr.String())
				}
				ser := *env.NewTSeries(addresses)
				return *env.NewBlock(ser)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//ip-addresses?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |email-addresses? |type? } 'block
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * block containing email addresses from Subject Alternative Names
	"x509-certificate//Email-addresses?": {
		Argsn: 1,
		Doc:   "Returns the email addresses from Subject Alternative Names of an X.509 certificate as a block of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//email-addresses?")
				}
				c := cert.Value.(*x509.Certificate)
				emails := make([]env.Object, len(c.EmailAddresses))
				for i, email := range c.EmailAddresses {
					emails[i] = *env.NewString(email)
				}
				ser := *env.NewTSeries(emails)
				return *env.NewBlock(ser)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//email-addresses?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |to-pem |type? } 'string
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * string containing the PEM-encoded certificate
	"x509-certificate//To-pem": {
		Argsn: 1,
		Doc:   "Converts an X.509 certificate to PEM-encoded string format.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//to-pem")
				}
				c := cert.Value.(*x509.Certificate)
				pemBlock := &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw}
				pemBytes := pem.EncodeToMemory(pemBlock)
				return *env.NewString(string(pemBytes))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//to-pem")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "key.pem" read-file |pem-decode |private-key-type? |type? } 'string
	// Args:
	// * private-key: private key as a native value
	// Returns:
	// * string containing the type of the private key (e.g., "*rsa.PrivateKey", "*ecdsa.PrivateKey")
	"private-key//Type?": {
		Argsn: 1,
		Doc:   "Returns the type of a private key as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch key := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(key.GetKind()) != "private-key" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "private-key//type?")
				}
				return *env.NewString(fmt.Sprintf("%T", key.Value))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "private-key//type?")
			}
		},
	},

	// Tests:
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |to-dict |type? } 'dict
	// ; equal { cc crypto "cert.pem" read-file |pem-decode |x509-parse-certificate |to-dict |get "Subject" |type? } 'string
	// Args:
	// * certificate: X.509 certificate as a native value
	// Returns:
	// * dictionary containing all certificate information
	"x509-certificate//To-dict": {
		Argsn: 1,
		Doc:   "Converts an X.509 certificate to a dictionary containing all certificate information for easy display and manipulation.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cert := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(cert.GetKind()) != "x509-certificate" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//to-dict")
				}
				c := cert.Value.(*x509.Certificate)

				// Create the main dictionary
				certDict := make(map[string]interface{})

				// Basic certificate information
				certDict["Subject"] = c.Subject.String()
				certDict["Issuer"] = c.Issuer.String()
				certDict["SerialNumber"] = c.SerialNumber.String()
				certDict["NotBefore"] = c.NotBefore.Format("2006-01-02 15:04:05 MST")
				certDict["NotAfter"] = c.NotAfter.Format("2006-01-02 15:04:05 MST")
				certDict["SignatureAlgorithm"] = c.SignatureAlgorithm.String()
				certDict["PublicKeyAlgorithm"] = c.PublicKeyAlgorithm.String()
				certDict["Version"] = int64(c.Version)
				certDict["IsCA"] = c.IsCA

				// Key Usage
				var keyUsages []interface{}
				if c.KeyUsage&x509.KeyUsageDigitalSignature > 0 {
					keyUsages = append(keyUsages, "Digital Signature")
				}
				if c.KeyUsage&x509.KeyUsageContentCommitment > 0 {
					keyUsages = append(keyUsages, "Content Commitment")
				}
				if c.KeyUsage&x509.KeyUsageKeyEncipherment > 0 {
					keyUsages = append(keyUsages, "Key Encipherment")
				}
				if c.KeyUsage&x509.KeyUsageDataEncipherment > 0 {
					keyUsages = append(keyUsages, "Data Encipherment")
				}
				if c.KeyUsage&x509.KeyUsageKeyAgreement > 0 {
					keyUsages = append(keyUsages, "Key Agreement")
				}
				if c.KeyUsage&x509.KeyUsageCertSign > 0 {
					keyUsages = append(keyUsages, "Cert Sign")
				}
				if c.KeyUsage&x509.KeyUsageCRLSign > 0 {
					keyUsages = append(keyUsages, "CRL Sign")
				}
				if c.KeyUsage&x509.KeyUsageEncipherOnly > 0 {
					keyUsages = append(keyUsages, "Encipher Only")
				}
				if c.KeyUsage&x509.KeyUsageDecipherOnly > 0 {
					keyUsages = append(keyUsages, "Decipher Only")
				}
				certDict["KeyUsage"] = keyUsages

				// Extended Key Usage
				var extKeyUsages []interface{}
				for _, extKeyUsage := range c.ExtKeyUsage {
					var usageStr string
					switch extKeyUsage {
					case x509.ExtKeyUsageServerAuth:
						usageStr = "Server Authentication"
					case x509.ExtKeyUsageClientAuth:
						usageStr = "Client Authentication"
					case x509.ExtKeyUsageCodeSigning:
						usageStr = "Code Signing"
					case x509.ExtKeyUsageEmailProtection:
						usageStr = "Email Protection"
					case x509.ExtKeyUsageTimeStamping:
						usageStr = "Time Stamping"
					case x509.ExtKeyUsageOCSPSigning:
						usageStr = "OCSP Signing"
					default:
						usageStr = fmt.Sprintf("Unknown (%d)", int(extKeyUsage))
					}
					extKeyUsages = append(extKeyUsages, usageStr)
				}
				certDict["ExtendedKeyUsage"] = extKeyUsages

				// Subject Alternative Names
				var dnsNames []interface{}
				for _, name := range c.DNSNames {
					dnsNames = append(dnsNames, name)
				}
				certDict["DNSNames"] = dnsNames

				var ipAddresses []interface{}
				for _, addr := range c.IPAddresses {
					ipAddresses = append(ipAddresses, addr.String())
				}
				certDict["IPAddresses"] = ipAddresses

				var emailAddresses []interface{}
				for _, email := range c.EmailAddresses {
					emailAddresses = append(emailAddresses, email)
				}
				certDict["EmailAddresses"] = emailAddresses

				// PEM encoded certificate
				pemBlock := &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw}
				pemBytes := pem.EncodeToMemory(pemBlock)
				certDict["PEM"] = string(pemBytes)

				// Expiration status
				certDict["IsExpired"] = time.Now().After(c.NotAfter)

				return *env.NewDict(certDict)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "x509-certificate//to-dict")
			}
		},
	},
}
