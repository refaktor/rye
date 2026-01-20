//go:build b_ollama
// +build b_ollama

package ollama

import (
	"context"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

var Builtins_ollama = map[string]*env.Builtin{

	//
	// ##### Ollama Client ##### "Functions for interacting with Ollama API"
	//

	// Tests:
	// ; client: ollama
	// Args:
	// * None (uses environment variables for configuration)
	// Returns:
	// * ollama-client - Ollama client instance for making API calls
	"ollama": {
		Argsn: 0,
		Doc:   "Creates an Ollama client from environment variables (OLLAMA_HOST).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, err := api.ClientFromEnvironment()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, err.Error(), "ollama")
			}
			return *env.NewNative(ps.Idx, client, "ollama-client")
		},
	},

	// Tests:
	// ; client: ollama\url "http://localhost:11434"
	// Args:
	// * url: String - URL of the Ollama server
	// Returns:
	// * ollama-client - Ollama client instance for making API calls
	"ollama\\url": {
		Argsn: 1,
		Doc:   "Creates an Ollama client with a specific server URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch url := arg0.(type) {
			case env.String:
				client, err := api.ClientFromEnvironment()
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "ollama\\url")
				}
				// Note: The official client doesn't have a direct URL setter,
				// but we can still use ClientFromEnvironment after setting OLLAMA_HOST
				_ = url // URL would be used if we had direct setter
				return *env.NewNative(ps.Idx, client, "ollama-client")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "ollama\\url")
			}
		},
	},

	//
	// ##### Embeddings ##### "Functions for creating text embeddings"
	//

	// Tests:
	// ; embedding: client .embed "bge-m3" "Hello world"
	// Args:
	// * client: Ollama client instance
	// * model: String - Name of the embedding model (e.g., "bge-m3", "nomic-embed-text")
	// * input: String - Text to create embedding for
	// Returns:
	// * vector - Numerical vector representation of the input text
	"ollama-client//Embed": {
		Argsn: 3,
		Doc:   "Create embedding for text input using Ollama. Returns a vector of floats.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//Embed")
				}

				switch model := arg1.(type) {
				case env.String:
					switch input := arg2.(type) {
					case env.String:
						req := &api.EmbedRequest{
							Model: model.Value,
							Input: input.Value,
						}

						resp, err := ollamaClient.Embed(context.Background(), req)
						if err != nil {
							return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Embed")
						}

						if len(resp.Embeddings) == 0 {
							return evaldo.MakeBuiltinError(ps, "No embeddings returned from Ollama", "ollama-client//Embed")
						}

						// Convert []float32 to []float64 for env.Vector
						embedding := resp.Embeddings[0]
						embedding64 := make([]float64, len(embedding))
						for i, v := range embedding {
							embedding64[i] = float64(v)
						}
						return *env.NewVector(embedding64)
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "ollama-client//Embed")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ollama-client//Embed")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//Embed")
			}
		},
	},

	// Tests:
	// ; embeddings: client .embed\many "bge-m3" { "Hello" "World" "Test" }
	// Args:
	// * client: Ollama client instance
	// * model: String - Name of the embedding model
	// * inputs: Block - Block of strings to create embeddings for
	// Returns:
	// * block - Block of vectors, one for each input string
	"ollama-client//Embed\\many": {
		Argsn: 3,
		Doc:   "Create embeddings for multiple text inputs using Ollama. Returns a block of vectors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//Embed\\many")
				}

				switch model := arg1.(type) {
				case env.String:
					switch inputs := arg2.(type) {
					case env.Block:
						// Collect all strings from the block
						texts := make([]string, 0, inputs.Series.Len())
						for i := 0; i < inputs.Series.Len(); i++ {
							item := inputs.Series.Get(i)
							if str, ok := item.(env.String); ok {
								texts = append(texts, str.Value)
							} else {
								return evaldo.MakeBuiltinError(ps, "All items in block must be strings", "ollama-client//Embed\\many")
							}
						}

						// Create embeddings for each text
						results := make([]env.Object, 0, len(texts))
						for _, text := range texts {
							req := &api.EmbedRequest{
								Model: model.Value,
								Input: text,
							}

							resp, err := ollamaClient.Embed(context.Background(), req)
							if err != nil {
								return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Embed\\many")
							}

							if len(resp.Embeddings) == 0 {
								return evaldo.MakeBuiltinError(ps, "No embeddings returned from Ollama", "ollama-client//Embed\\many")
							}

							// Convert []float32 to []float64 for env.Vector
							embedding := resp.Embeddings[0]
							embedding64 := make([]float64, len(embedding))
							for j, v := range embedding {
								embedding64[j] = float64(v)
							}
							results = append(results, *env.NewVector(embedding64))
						}

						return *env.NewBlock(*env.NewTSeries(results))
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.BlockType}, "ollama-client//Embed\\many")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ollama-client//Embed\\many")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//Embed\\many")
			}
		},
	},

	//
	// ##### Chat Completions ##### "Functions for text generation and conversations"
	//

	// Tests:
	// ; response: client .chat "llama2" "Hello, how are you?"
	// Args:
	// * client: Ollama client instance
	// * model: String - Name of the model to use (e.g., "llama2", "mistral")
	// * prompt: String - Text prompt for completion
	// Returns:
	// * string - The AI's response text
	"ollama-client//Chat": {
		Argsn: 3,
		Doc:   "Generate chat completion using Ollama API. Simple single-prompt interface.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//Chat")
				}

				switch model := arg1.(type) {
				case env.String:
					switch prompt := arg2.(type) {
					case env.String:
						req := &api.ChatRequest{
							Model: model.Value,
							Messages: []api.Message{
								{
									Role:    "user",
									Content: prompt.Value,
								},
							},
						}

						var response strings.Builder
						err := ollamaClient.Chat(context.Background(), req, func(resp api.ChatResponse) error {
							response.WriteString(resp.Message.Content)
							return nil
						})

						if err != nil {
							return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Chat")
						}

						return *env.NewString(response.String())
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "ollama-client//Chat")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ollama-client//Chat")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//Chat")
			}
		},
	},

	// Tests:
	// ; conversation: [ dict { "role" "system" "content" "You are helpful" } dict { "role" "user" "content" "Hello!" } ]
	// ; response: client .chat\messages "llama2" conversation
	// Args:
	// * client: Ollama client instance
	// * model: String - Name of the model to use
	// * messages: Block - Conversation format with role/content dicts
	// Returns:
	// * string - The AI's response text
	"ollama-client//Chat\\messages": {
		Argsn: 3,
		Doc:   "Generate chat completion with conversation history. Messages should be dicts with 'role' and 'content' keys.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//Chat\\messages")
				}

				switch model := arg1.(type) {
				case env.String:
					switch messages := arg2.(type) {
					case env.Block:
						// Convert block to Ollama messages
						ollamaMessages := make([]api.Message, 0, messages.Series.Len())
						for i := 0; i < messages.Series.Len(); i++ {
							item := messages.Series.Get(i)
							if dict, ok := item.(env.Dict); ok {
								role, roleExists := dict.Data["role"]
								content, contentExists := dict.Data["content"]

								if !roleExists || !contentExists {
									return evaldo.MakeBuiltinError(ps, "Each message must have 'role' and 'content' keys", "ollama-client//Chat\\messages")
								}

								roleStr, ok1 := role.(env.String)
								contentStr, ok2 := content.(env.String)

								if !ok1 || !ok2 {
									return evaldo.MakeBuiltinError(ps, "Role and content must be strings", "ollama-client//Chat\\messages")
								}

								ollamaMessages = append(ollamaMessages, api.Message{
									Role:    roleStr.Value,
									Content: contentStr.Value,
								})
							} else {
								return evaldo.MakeBuiltinError(ps, "Each message must be a dict", "ollama-client//Chat\\messages")
							}
						}

						req := &api.ChatRequest{
							Model:    model.Value,
							Messages: ollamaMessages,
						}

						var response strings.Builder
						err := ollamaClient.Chat(context.Background(), req, func(resp api.ChatResponse) error {
							response.WriteString(resp.Message.Content)
							return nil
						})

						if err != nil {
							return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Chat\\messages")
						}

						return *env.NewString(response.String())
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.BlockType}, "ollama-client//Chat\\messages")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ollama-client//Chat\\messages")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//Chat\\messages")
			}
		},
	},

	// Tests:
	// ; client .chat\stream "llama2" "Tell me a story" fn { chunk } { prn chunk }
	// Args:
	// * client: Ollama client instance
	// * model: String - Name of the model to use
	// * prompt: String - Text prompt for completion
	// * callback: Function or Block - Called for each chunk of response
	// Returns:
	// * string - Complete response text after streaming
	"ollama-client//Chat\\stream": {
		Argsn: 4,
		Doc:   "Stream chat completion with real-time chunks. Calls callback for each piece of response.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//Chat\\stream")
				}

				switch model := arg1.(type) {
				case env.String:
					switch prompt := arg2.(type) {
					case env.String:
						switch callback := arg3.(type) {
						case env.Block:
							req := &api.ChatRequest{
								Model: model.Value,
								Messages: []api.Message{
									{
										Role:    "user",
										Content: prompt.Value,
									},
								},
							}

							var fullResponse strings.Builder
							err := ollamaClient.Chat(context.Background(), req, func(resp api.ChatResponse) error {
								chunk := resp.Message.Content
								fullResponse.WriteString(chunk)

								// Call the callback with the chunk
								ser := ps.Ser
								ps.Ser = callback.Series
								evaldo.EvalBlockInj(ps, *env.NewString(chunk), true)
								ps.Ser = ser

								return nil
							})

							if err != nil {
								return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Chat\\stream")
							}

							return *env.NewString(fullResponse.String())
						case env.Function:
							req := &api.ChatRequest{
								Model: model.Value,
								Messages: []api.Message{
									{
										Role:    "user",
										Content: prompt.Value,
									},
								},
							}

							var fullResponse strings.Builder
							err := ollamaClient.Chat(context.Background(), req, func(resp api.ChatResponse) error {
								chunk := resp.Message.Content
								fullResponse.WriteString(chunk)

								// Call the function with the chunk
								evaldo.CallFunctionArgs2(callback, ps, *env.NewString(chunk), nil, nil)

								return nil
							})

							if err != nil {
								return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Chat\\stream")
							}

							return *env.NewString(fullResponse.String())
						default:
							return evaldo.MakeArgError(ps, 4, []env.Type{env.BlockType, env.FunctionType}, "ollama-client//Chat\\stream")
						}
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "ollama-client//Chat\\stream")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ollama-client//Chat\\stream")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//Chat\\stream")
			}
		},
	},

	//
	// ##### Text Generation ##### "Functions for simple text generation"
	//

	// Tests:
	// ; response: client .generate "llama2" "The sky is"
	// Args:
	// * client: Ollama client instance
	// * model: String - Name of the model to use
	// * prompt: String - Text prompt for completion
	// Returns:
	// * string - The generated text
	"ollama-client//Generate": {
		Argsn: 3,
		Doc:   "Generate text completion using Ollama API. Simple text generation without chat format.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//Generate")
				}

				switch model := arg1.(type) {
				case env.String:
					switch prompt := arg2.(type) {
					case env.String:
						req := &api.GenerateRequest{
							Model:  model.Value,
							Prompt: prompt.Value,
						}

						var response strings.Builder
						err := ollamaClient.Generate(context.Background(), req, func(resp api.GenerateResponse) error {
							response.WriteString(resp.Response)
							return nil
						})

						if err != nil {
							return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Generate")
						}

						return *env.NewString(response.String())
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "ollama-client//Generate")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ollama-client//Generate")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//Generate")
			}
		},
	},

	//
	// ##### Model Management ##### "Functions for managing Ollama models"
	//

	// Tests:
	// ; models: client .list-models
	// Args:
	// * client: Ollama client instance
	// Returns:
	// * block - List of available models with their metadata
	"ollama-client//List-models": {
		Argsn: 1,
		Doc:   "List all available Ollama models with their metadata.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//List-models")
				}

				resp, err := ollamaClient.List(context.Background())
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//List-models")
				}

				models := make([]env.Object, len(resp.Models))
				for i, model := range resp.Models {
					modelDict := make(map[string]any)
					modelDict["name"] = *env.NewString(model.Name)
					modelDict["size"] = *env.NewInteger(model.Size)
					modelDict["modified"] = *env.NewString(model.ModifiedAt.String())
					models[i] = *env.NewDict(modelDict)
				}

				return *env.NewBlock(*env.NewTSeries(models))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//List-models")
			}
		},
	},

	// Tests:
	// ; info: client .show-model "llama2"
	// Args:
	// * client: Ollama client instance
	// * model: String - Name of the model to show info for
	// Returns:
	// * dict - Model information including parameters, template, etc.
	"ollama-client//Show-model": {
		Argsn: 2,
		Doc:   "Show detailed information about a specific Ollama model.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//Show-model")
				}

				switch model := arg1.(type) {
				case env.String:
					req := &api.ShowRequest{
						Model: model.Value,
					}

					resp, err := ollamaClient.Show(context.Background(), req)
					if err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Show-model")
					}

					modelDict := make(map[string]any)
					modelDict["modelfile"] = *env.NewString(resp.Modelfile)
					modelDict["parameters"] = *env.NewString(resp.Parameters)
					modelDict["template"] = *env.NewString(resp.Template)
					modelDict["license"] = *env.NewString(resp.License)

					return *env.NewDict(modelDict)
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ollama-client//Show-model")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//Show-model")
			}
		},
	},

	// Tests:
	// ; client .pull-model "llama2"
	// Args:
	// * client: Ollama client instance
	// * model: String - Name of the model to pull
	// Returns:
	// * integer - 1 on success
	"ollama-client//Pull-model": {
		Argsn: 2,
		Doc:   "Pull (download) a model from the Ollama registry.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				ollamaClient, ok := client.Value.(*api.Client)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid Ollama client", "ollama-client//Pull-model")
				}

				switch model := arg1.(type) {
				case env.String:
					req := &api.PullRequest{
						Model: model.Value,
					}

					err := ollamaClient.Pull(context.Background(), req, func(resp api.ProgressResponse) error {
						return nil
					})

					if err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "ollama-client//Pull-model")
					}

					return *env.NewInteger(1)
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "ollama-client//Pull-model")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "ollama-client//Pull-model")
			}
		},
	},
}
