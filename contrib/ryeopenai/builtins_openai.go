//go:build b_openai
// +build b_openai

package ryeopenai

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"

	"github.com/drewlanenga/govector"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	openai "github.com/sashabaranov/go-openai"
)

var Builtins_openai = map[string]*env.Builtin{

	"new-openai-client": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch mpi := arg0.(type) {
			case env.String:
				client := openai.NewClient(mpi.Value)
				//client := openai.newClient()
				return *env.NewNative(ps.Idx, client, "openai-client")
			default:
				return evaldo.MakeError(ps, "Arg 1 not string.")
			}
		},
	},

	"openai-client//complete-chat": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.Native:
				switch s := arg1.(type) {
				case env.String:
					client := c.Value.(*openai.Client)
					resp, err := client.CreateChatCompletion(
						context.Background(),
						openai.ChatCompletionRequest{
							Model: openai.GPT3Dot5Turbo,
							Messages: []openai.ChatCompletionMessage{
								{
									Role:    openai.ChatMessageRoleUser,
									Content: s.Value,
								},
							},
						},
					)

					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}
					return env.String{resp.Choices[0].Message.Content}
				default:
					return evaldo.MakeError(ps, "Arg 2 not string.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},
	"openai-client//create-embeddings": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.Native:
				switch s := arg1.(type) {
				case env.String:
					val := make([]string, 1)
					val[0] = s.Value
					client := c.Value.(*openai.Client)
					resp, err := client.CreateEmbeddings(
						context.Background(),
						openai.EmbeddingRequest{
							User:  "Rye-demo",
							Input: val,
							Model: openai.AdaEmbeddingV2,
						},
					)
					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}
					val2, err2 := govector.AsVector(resp.Data[0].Embedding)
					if err2 != nil {
						return evaldo.MakeError(ps, err2.Error())
					}
					return *env.NewVector(val2)
					//return *env.NewNative(ps.Idx, , "vector")
				default:
					return evaldo.MakeError(ps, "Arg 2 not string.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},

	"openai-embedding//to-bytes": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.Native:
				buf := new(bytes.Buffer)
				err := binary.Write(buf, binary.LittleEndian, c.Value.([]float32))
				if err != nil {
					return evaldo.MakeError(ps, err.Error())
				}
				fmt.Printf("% x", buf.Bytes())
				return *env.NewNative(ps.Idx, buf.Bytes, "bytes")
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},
}
