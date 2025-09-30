//go:build b_openai
// +build b_openai

package ryeopenai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

var Builtins_openai = map[string]*env.Builtin{

	//
	// ##### OpenAI Client ##### "Functions for interacting with OpenAI API"
	//

	// Tests:
	// client: openai "sk-your-api-key"
	// Args:
	// * api-key: String - Your OpenAI API key
	// Returns:
	// * openai-client - OpenAI client instance for making API calls
	"openai": {
		Argsn: 1,
		Doc:   "Creates an OpenAI client with the provided API key.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch apiKey := arg0.(type) {
			case env.String:
				client := openai.NewClient(option.WithAPIKey(apiKey.Value))
				return *env.NewNative(ps.Idx, client, "openai-client")
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a string (API key).")
			}
		},
	},

	//
	// ##### Chat Completions ##### "Functions for text generation and conversations"
	//

	// Tests:
	// response: client .chat "Hello, how are you?"
	// conversation: [ { "role" "system" "content" "You are a helpful assistant" } { "role" "user" "content" "Hello!" } ]
	// response2: client .chat conversation
	// Args:
	// * client: OpenAI client instance
	// * prompt: String (simple prompt) or Block (conversation format with role/content dicts)
	// Returns:
	// * string - The AI's response text
	"openai-client//Chat": {
		Argsn: 2,
		Doc:   "Generate chat completion using OpenAI API. Accepts either a simple string prompt or a conversation format with message history.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					openaiClient := client.Value.(openai.Client)

					response, err := openaiClient.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
						Messages: []openai.ChatCompletionMessageParamUnion{
							openai.UserMessage(prompt.Value),
						},
						Model: openai.ChatModelGPT4oMini,
					})

					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}

					if len(response.Choices) == 0 {
						return evaldo.MakeError(ps, "No response choices returned from OpenAI")
					}

					return *env.NewString(response.Choices[0].Message.Content)
				case env.Block:
					// Handle conversation format: [ { "role" "user" "content" "Hello" } { "role" "assistant" "content" "Hi!" } ]
					openaiClient := client.Value.(openai.Client)

					var messages []openai.ChatCompletionMessageParamUnion

					for i, item := range prompt.Series.S {
						if dict, ok := item.(env.Dict); ok {
							role, roleExists := dict.Data["role"]
							content, contentExists := dict.Data["content"]

							if !roleExists || !contentExists {
								return evaldo.MakeError(ps, fmt.Sprintf("Message at index %d must have 'role' and 'content' fields", i))
							}

							roleStr, ok1 := role.(env.String)
							contentStr, ok2 := content.(env.String)

							if !ok1 || !ok2 {
								return evaldo.MakeError(ps, fmt.Sprintf("Message at index %d: 'role' and 'content' must be strings", i))
							}

							switch roleStr.Value {
							case "user":
								messages = append(messages, openai.UserMessage(contentStr.Value))
							case "assistant":
								messages = append(messages, openai.AssistantMessage(contentStr.Value))
							case "system":
								messages = append(messages, openai.SystemMessage(contentStr.Value))
							default:
								return evaldo.MakeError(ps, fmt.Sprintf("Invalid role '%s' at index %d. Must be 'user', 'assistant', or 'system'", roleStr.Value, i))
							}
						} else {
							return evaldo.MakeError(ps, fmt.Sprintf("Message at index %d must be a dictionary", i))
						}
					}

					response, err := openaiClient.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
						Messages: messages,
						Model:    openai.ChatModelGPT4oMini,
					})

					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}

					if len(response.Choices) == 0 {
						return evaldo.MakeError(ps, "No response choices returned from OpenAI")
					}

					return *env.NewString(response.Choices[0].Message.Content)
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt) or block (conversation).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	// Tests:
	// options: { "model" "gpt-4" "temperature" 0.7 "max-tokens" 150 }
	// response: client .chat\opts "Hello!" options
	// Args:
	// * client: OpenAI client instance
	// * prompt: String - Text prompt for completion
	// * options: Dict - Configuration options (model, temperature, max-tokens)
	// Returns:
	// * string - The AI's response text
	"openai-client//Chat\\opts": {
		Argsn: 3,
		Doc:   "Generate chat completion with custom options like model, temperature, and max-tokens.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					switch options := arg2.(type) {
					case env.Dict:
						openaiClient := client.Value.(openai.Client)

						params := openai.ChatCompletionNewParams{
							Messages: []openai.ChatCompletionMessageParamUnion{
								openai.UserMessage(prompt.Value),
							},
							Model: openai.ChatModelGPT4oMini,
						}

						// Process options
						if model, exists := options.Data["model"]; exists {
							if modelStr, ok := model.(env.String); ok {
								params.Model = openai.ChatModel(modelStr.Value)
							}
						}

						if temp, exists := options.Data["temperature"]; exists {
							if tempFloat, ok := temp.(env.Decimal); ok {
								params.Temperature = openai.Float(tempFloat.Value)
							}
						}

						if maxTokens, exists := options.Data["max-tokens"]; exists {
							if maxTokensInt, ok := maxTokens.(env.Integer); ok {
								params.MaxTokens = openai.Int(int64(maxTokensInt.Value))
							}
						}

						response, err := openaiClient.Chat.Completions.New(context.Background(), params)

						if err != nil {
							return evaldo.MakeError(ps, err.Error())
						}

						if len(response.Choices) == 0 {
							return evaldo.MakeError(ps, "No response choices returned from OpenAI")
						}

						return *env.NewString(response.Choices[0].Message.Content)
					default:
						return evaldo.MakeError(ps, "Arg 3 must be a dictionary (options).")
					}
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	//
	// ##### Embeddings ##### "Functions for creating text embeddings"
	//

	// Tests:
	// ; embeddings: client .create-embeddings "sample text"
	// Args:
	// * client: OpenAI client instance
	// * input: String or Block - Text(s) to create embeddings for
	// Returns:
	// * block - Numerical vector representations of the input text(s)
	"openai-client//Create-embeddings": {
		Argsn: 2,
		Doc:   "Create embeddings for text input. Currently not fully implemented - placeholder function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0.(type) {
			case env.Native:
				switch arg1.(type) {
				case env.String:
					// Placeholder for embeddings functionality
					return evaldo.MakeError(ps, "Embeddings functionality not yet fully implemented with official OpenAI client - API compatibility issues")
				case env.Block:
					// Placeholder for embeddings functionality
					return evaldo.MakeError(ps, "Embeddings functionality not yet fully implemented with official OpenAI client - API compatibility issues")
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string or block of strings.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	//
	// ##### Image Generation ##### "Functions for generating images with DALL-E"
	//

	// Tests:
	// image-url: client .generate-image "A red bicycle in a park"
	// Args:
	// * client: OpenAI client instance
	// * prompt: String - Description of the image to generate
	// Returns:
	// * string - URL of the generated image
	"openai-client//Generate-image": {
		Argsn: 2,
		Doc:   "Generate an image using DALL-E based on a text prompt. Returns URL of the generated image.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					openaiClient := client.Value.(openai.Client)

					response, err := openaiClient.Images.Generate(context.Background(), openai.ImageGenerateParams{
						Prompt:         prompt.Value,
						Model:          openai.ImageModelDallE3,
						Size:           openai.ImageGenerateParamsSize1024x1024,
						Quality:        openai.ImageGenerateParamsQualityStandard,
						ResponseFormat: openai.ImageGenerateParamsResponseFormatURL,
					})

					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}

					if len(response.Data) == 0 {
						return evaldo.MakeError(ps, "No image data returned from OpenAI")
					}

					// Return the image URL
					return *env.NewString(response.Data[0].URL)
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	// Tests:
	// options: { "size" "512x512" "quality" "hd" "model" "dall-e-3" }
	// image-url: client .generate-image\opts "A blue cat" options
	// Args:
	// * client: OpenAI client instance
	// * prompt: String - Description of the image to generate
	// * options: Dict - Image generation options (size, quality, model)
	// Returns:
	// * string - URL of the generated image
	"openai-client//Generate-image\\opts": {
		Argsn: 3,
		Doc:   "Generate an image using DALL-E with custom options like size, quality, and model.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					switch options := arg2.(type) {
					case env.Dict:
						openaiClient := client.Value.(openai.Client)

						params := openai.ImageGenerateParams{
							Prompt:         prompt.Value,
							Model:          openai.ImageModelDallE3,
							Size:           openai.ImageGenerateParamsSize1024x1024,
							Quality:        openai.ImageGenerateParamsQualityStandard,
							ResponseFormat: openai.ImageGenerateParamsResponseFormatURL,
						}

						// Process options
						if size, exists := options.Data["size"]; exists {
							if sizeStr, ok := size.(env.String); ok {
								switch sizeStr.Value {
								case "256x256":
									params.Size = openai.ImageGenerateParamsSize256x256
								case "512x512":
									params.Size = openai.ImageGenerateParamsSize512x512
								case "1024x1024":
									params.Size = openai.ImageGenerateParamsSize1024x1024
								case "1792x1024":
									params.Size = openai.ImageGenerateParamsSize1792x1024
								case "1024x1792":
									params.Size = openai.ImageGenerateParamsSize1024x1792
								}
							}
						}

						if quality, exists := options.Data["quality"]; exists {
							if qualityStr, ok := quality.(env.String); ok {
								switch qualityStr.Value {
								case "standard":
									params.Quality = openai.ImageGenerateParamsQualityStandard
								case "hd":
									params.Quality = openai.ImageGenerateParamsQualityHD
								}
							}
						}

						if model, exists := options.Data["model"]; exists {
							if modelStr, ok := model.(env.String); ok {
								switch modelStr.Value {
								case "dall-e-2":
									params.Model = openai.ImageModelDallE2
								case "dall-e-3":
									params.Model = openai.ImageModelDallE3
								}
							}
						}

						response, err := openaiClient.Images.Generate(context.Background(), params)

						if err != nil {
							return evaldo.MakeError(ps, err.Error())
						}

						if len(response.Data) == 0 {
							return evaldo.MakeError(ps, "No image data returned from OpenAI")
						}

						// Return the image URL
						return *env.NewString(response.Data[0].URL)
					default:
						return evaldo.MakeError(ps, "Arg 3 must be a dictionary (options).")
					}
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	//
	// ##### Models and Utilities ##### "Functions for model information and utilities"
	//

	// Tests:
	// models: client .list-models
	// Args:
	// * client: OpenAI client instance
	// Returns:
	// * block - List of available models with their metadata
	"openai-client//List-models": {
		Argsn: 1,
		Doc:   "List all available OpenAI models with their metadata (id, created date, owner).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				openaiClient := client.Value.(openai.Client)

				response, err := openaiClient.Models.List(context.Background())

				if err != nil {
					return evaldo.MakeError(ps, err.Error())
				}

				// Convert models to Rye block
				models := make([]env.Object, len(response.Data))

				for i, model := range response.Data {
					modelDict := make(map[string]any)
					modelDict["id"] = *env.NewString(model.ID)
					modelDict["created"] = *env.NewInteger(int64(model.Created))
					modelDict["owned-by"] = *env.NewString(model.OwnedBy)

					models[i] = *env.NewDict(modelDict)
				}

				return *env.NewBlock(*env.NewTSeries(models))
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	// Tests:
	// ; text: client .transcribe-audio "/path/to/audio.mp3"
	// Args:
	// * client: OpenAI client instance
	// * file-path: String - Path to audio file for transcription
	// Returns:
	// * string - Transcribed text from the audio file
	"openai-client//Transcribe-audio": {
		Argsn: 2,
		Doc:   "Transcribe audio file to text using OpenAI Whisper. Currently not fully implemented - placeholder function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0.(type) {
			case env.Native:
				switch filePath := arg1.(type) {
				case env.String:
					// This is a placeholder implementation
					// In practice, you would need to open and read the file properly
					// For now, return an error indicating this needs proper file handling
					return evaldo.MakeError(ps, "Audio transcription not yet fully implemented - requires proper file handling for: "+filePath.Value)
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (file path).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	// Tests:
	// ; json-data: client .get-response-json "Return JSON: {\"name\": \"test\", \"value\": 42}"
	// Args:
	// * client: OpenAI client instance
	// * prompt: String - Prompt requesting JSON response
	// Returns:
	// * dict or block - Parsed JSON response as Rye objects
	"openai-client//Get-response-json": {
		Argsn: 2,
		Doc:   "Get a chat response and parse it as JSON, converting to Rye objects (dicts, blocks, etc.).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					openaiClient := client.Value.(openai.Client)

					response, err := openaiClient.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
						Messages: []openai.ChatCompletionMessageParamUnion{
							openai.UserMessage(prompt.Value),
						},
						Model: openai.ChatModelGPT4oMini,
						// Note: JSON response format may need to be handled differently with the official client
					})

					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}

					if len(response.Choices) == 0 {
						return evaldo.MakeError(ps, "No response choices returned from OpenAI")
					}

					// Parse JSON response
					var jsonData interface{}
					err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &jsonData)
					if err != nil {
						return evaldo.MakeError(ps, fmt.Sprintf("Failed to parse JSON response: %s", err.Error()))
					}

					// Convert JSON to Rye object
					ryeObj, err := jsonToRye(jsonData, ps.Idx)
					if err != nil {
						return evaldo.MakeError(ps, fmt.Sprintf("Failed to convert JSON to Rye object: %s", err.Error()))
					}

					return ryeObj
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	//
	// ##### Streaming Chat ##### "Functions for streaming chat completions"
	//

	// Tests:
	// ; client .chat\stream "Tell a story" { |chunk| prn chunk }
	// Args:
	// * client: OpenAI client instance
	// * prompt: String or Block - Text prompt or conversation format
	// * callback: Block - Function to call for each chunk of streamed response
	// Returns:
	// * string - Complete response text after streaming is finished
	"openai-client//Chat\\stream": {
		Argsn: 3,
		Doc:   "Stream chat completion with real-time chunks. Calls callback function for each piece of the response as it arrives.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					switch callback := arg2.(type) {
					case env.Block:
						openaiClient := client.Value.(openai.Client)

						stream := openaiClient.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
							Messages: []openai.ChatCompletionMessageParamUnion{
								openai.UserMessage(prompt.Value),
							},
							Model: openai.ChatModelGPT4oMini,
						})

						var fullResponse strings.Builder
						for stream.Next() {
							chunk := stream.Current()
							if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
								content := chunk.Choices[0].Delta.Content
								fullResponse.WriteString(content)

								// Call the callback with the chunk
								// Create a new environment and evaluate the callback block
								ser := ps.Ser
								ps.Ser = callback.Series
								evaldo.EvalBlockInj(ps, *env.NewString(content), true)
								ps.Ser = ser

								if ps.ErrorFlag {
									stream.Close()
									return ps.Res
								}
							}
						}

						if err := stream.Err(); err != nil {
							return evaldo.MakeError(ps, err.Error())
						}

						// Return the full response
						return *env.NewString(fullResponse.String())
					default:
						return evaldo.MakeError(ps, "Arg 3 must be a block (callback function).")
					}
				case env.Block:
					// Handle conversation format with streaming
					switch callback := arg2.(type) {
					case env.Block:
						openaiClient := client.Value.(openai.Client)

						var messages []openai.ChatCompletionMessageParamUnion

						for i, item := range prompt.Series.S {
							if dict, ok := item.(env.Dict); ok {
								role, roleExists := dict.Data["role"]
								content, contentExists := dict.Data["content"]

								if !roleExists || !contentExists {
									return evaldo.MakeError(ps, fmt.Sprintf("Message at index %d must have 'role' and 'content' fields", i))
								}

								roleStr, ok1 := role.(env.String)
								contentStr, ok2 := content.(env.String)

								if !ok1 || !ok2 {
									return evaldo.MakeError(ps, fmt.Sprintf("Message at index %d: 'role' and 'content' must be strings", i))
								}

								switch roleStr.Value {
								case "user":
									messages = append(messages, openai.UserMessage(contentStr.Value))
								case "assistant":
									messages = append(messages, openai.AssistantMessage(contentStr.Value))
								case "system":
									messages = append(messages, openai.SystemMessage(contentStr.Value))
								default:
									return evaldo.MakeError(ps, fmt.Sprintf("Invalid role '%s' at index %d. Must be 'user', 'assistant', or 'system'", roleStr.Value, i))
								}
							} else {
								return evaldo.MakeError(ps, fmt.Sprintf("Message at index %d must be a dictionary", i))
							}
						}

						stream := openaiClient.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
							Messages: messages,
							Model:    openai.ChatModelGPT4oMini,
						})

						var fullResponse strings.Builder
						for stream.Next() {
							chunk := stream.Current()
							if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
								content := chunk.Choices[0].Delta.Content
								fullResponse.WriteString(content)

								// Call the callback with the chunk
								// Create a new environment and evaluate the callback block
								ser := ps.Ser
								ps.Ser = callback.Series
								evaldo.EvalBlockInj(ps, *env.NewString(content), true)
								ps.Ser = ser

								if ps.ErrorFlag {
									stream.Close()
									return ps.Res
								}
							}
						}

						if err := stream.Err(); err != nil {
							return evaldo.MakeError(ps, err.Error())
						}

						// Return the full response
						return *env.NewString(fullResponse.String())
					default:
						return evaldo.MakeError(ps, "Arg 3 must be a block (callback function).")
					}
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt) or block (conversation).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},

	// Tests:
	// ; options: { "model" "gpt-4" "temperature" 0.5 }
	// ; client .chat\stream\opts "Tell a story" options { |chunk| prn chunk }
	// Args:
	// * client: OpenAI client instance
	// * prompt: String - Text prompt for completion
	// * options: Dict - Configuration options (model, temperature, max-tokens)
	// * callback: Block - Function to call for each chunk of streamed response
	// Returns:
	// * string - Complete response text after streaming is finished
	"openai-client//Chat\\stream\\opts": {
		Argsn: 4,
		Doc:   "Stream chat completion with custom options and real-time chunks. Combines streaming with configuration control.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					switch options := arg2.(type) {
					case env.Dict:
						switch callback := arg3.(type) {
						case env.Block:
							openaiClient := client.Value.(openai.Client)

							params := openai.ChatCompletionNewParams{
								Messages: []openai.ChatCompletionMessageParamUnion{
									openai.UserMessage(prompt.Value),
								},
								Model: openai.ChatModelGPT4oMini,
							}

							// Process options
							if model, exists := options.Data["model"]; exists {
								if modelStr, ok := model.(env.String); ok {
									params.Model = openai.ChatModel(modelStr.Value)
								}
							}

							if temp, exists := options.Data["temperature"]; exists {
								if tempFloat, ok := temp.(env.Decimal); ok {
									params.Temperature = openai.Float(tempFloat.Value)
								}
							}

							if maxTokens, exists := options.Data["max-tokens"]; exists {
								if maxTokensInt, ok := maxTokens.(env.Integer); ok {
									params.MaxTokens = openai.Int(int64(maxTokensInt.Value))
								}
							}

							stream := openaiClient.Chat.Completions.NewStreaming(context.Background(), params)

							var fullResponse strings.Builder
							for stream.Next() {
								chunk := stream.Current()
								if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
									content := chunk.Choices[0].Delta.Content
									fullResponse.WriteString(content)

									// Call the callback with the chunk
									// Create a new environment and evaluate the callback block
									ser := ps.Ser
									ps.Ser = callback.Series
									evaldo.EvalBlockInj(ps, *env.NewString(content), true)
									ps.Ser = ser

									if ps.ErrorFlag {
										stream.Close()
										return ps.Res
									}
								}
							}

							if err := stream.Err(); err != nil {
								return evaldo.MakeError(ps, err.Error())
							}

							// Return the full response
							return *env.NewString(fullResponse.String())
						default:
							return evaldo.MakeError(ps, "Arg 4 must be a block (callback function).")
						}
					default:
						return evaldo.MakeError(ps, "Arg 3 must be a dictionary (options).")
					}
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be an OpenAI client.")
			}
		},
	},
}

// Helper function to convert JSON to Rye objects
func jsonToRye(data interface{}, idx *env.Idxs) (env.Object, error) {
	switch v := data.(type) {
	case string:
		return *env.NewString(v), nil
	case float64:
		return *env.NewDecimal(v), nil
	case int64:
		return *env.NewInteger(v), nil
	case bool:
		if v {
			return *env.NewInteger(1), nil
		}
		return *env.NewInteger(0), nil
	case nil:
		return *env.NewString(""), nil
	case map[string]interface{}:
		dict := make(map[string]any)
		for k, v := range v {
			ryeObj, err := jsonToRye(v, idx)
			if err != nil {
				return *env.NewString(""), err
			}
			dict[k] = ryeObj
		}
		return *env.NewDict(dict), nil
	case []interface{}:
		objs := make([]env.Object, len(v))
		for i, item := range v {
			ryeObj, err := jsonToRye(item, idx)
			if err != nil {
				return *env.NewString(""), err
			}
			objs[i] = ryeObj
		}
		return *env.NewBlock(*env.NewTSeries(objs)), nil
	default:
		return *env.NewString(fmt.Sprintf("%v", v)), nil
	}
}
