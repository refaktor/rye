//go:build b_genai
// +build b_genai

package genai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"google.golang.org/genai"
)

var Builtins_genai = map[string]*env.Builtin{

	//
	// ##### Google GenAI Client ##### "Functions for interacting with Google GenAI (Gemini) API"
	//

	// Tests:
	// client: genai
	// Args:
	// * None (uses GEMINI_API_KEY env var)
	// Returns:
	// * genai-client - Google GenAI client instance for making API calls
	"genai": {
		Argsn: 0,
		Doc:   "Creates a Google GenAI client using the GEMINI_API_KEY environment variable.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ctx := context.Background()
			client, err := genai.NewClient(ctx, nil)
			if err != nil {
				return evaldo.MakeError(ps, fmt.Sprintf("Failed to initialize GenAI client: %v", err))
			}
			return *env.NewNative(ps.Idx, client, "genai-client")
		},
	},

	// Tests:
	// client: genai\key "your-api-key"
	// Args:
	// * api-key: String - Your Google GenAI API key
	// Returns:
	// * genai-client - Google GenAI client instance for making API calls
	"genai\\key": {
		Argsn: 1,
		Doc:   "Creates a Google GenAI client with the provided API key.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch apiKey := arg0.(type) {
			case env.String:
				// Set the API key as environment variable
				os.Setenv("GEMINI_API_KEY", apiKey.Value)
				ctx := context.Background()
				client, err := genai.NewClient(ctx, nil)
				if err != nil {
					return evaldo.MakeError(ps, fmt.Sprintf("Failed to initialize GenAI client: %v", err))
				}
				return *env.NewNative(ps.Idx, client, "genai-client")
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a string (API key).")
			}
		},
	},

	//
	// ##### Text Generation ##### "Functions for text generation with Gemini models"
	//

	// Tests:
	// response: client .generate "Hello, how are you?"
	// Args:
	// * client: GenAI client instance
	// * prompt: String - Text prompt for generation
	// Returns:
	// * string - The AI's response text
	"genai-client//Generate": {
		Argsn: 2,
		Doc:   "Generate text using Google GenAI. Uses gemini-2.5-flash model by default.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					genaiClient := client.Value.(*genai.Client)

					content := []*genai.Content{
						{Parts: []*genai.Part{genai.NewPartFromText(prompt.Value)}},
					}

					resp, err := genaiClient.Models.GenerateContent(context.Background(), "gemini-2.5-flash", content, nil)
					if err != nil {
						return evaldo.MakeError(ps, fmt.Sprintf("Generation failed: %v", err))
					}

					responseText := resp.Text()
					if responseText == "" {
						return evaldo.MakeError(ps, "No response text returned from GenAI")
					}

					return *env.NewString(responseText)
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a GenAI client.")
			}
		},
	},

	// Tests:
	// options: { "model" "gemini-pro" "temperature" 0.7 }
	// response: client .generate\opts "Hello!" options
	// Args:
	// * client: GenAI client instance
	// * prompt: String - Text prompt for generation
	// * options: Dict - Configuration options (model, temperature)
	// Returns:
	// * string - The AI's response text
	"genai-client//Generate\\opts": {
		Argsn: 3,
		Doc:   "Generate text with custom options like model and temperature.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					switch options := arg2.(type) {
					case env.Dict:
						genaiClient := client.Value.(*genai.Client)

						// Default model
						model := "gemini-2.5-flash"

						// Default config
						config := &genai.GenerateContentConfig{}

						// Process options
						if modelOpt, exists := options.Data["model"]; exists {
							if modelStr, ok := modelOpt.(env.String); ok {
								model = modelStr.Value
							}
						}

						if temp, exists := options.Data["temperature"]; exists {
							if tempFloat, ok := temp.(env.Decimal); ok {
								config.Temperature = genai.Ptr(float32(tempFloat.Value))
							}
						}

						if maxTokens, exists := options.Data["max-tokens"]; exists {
							if maxTokensInt, ok := maxTokens.(env.Integer); ok {
								config.MaxOutputTokens = int32(maxTokensInt.Value)
							}
						}

						content := []*genai.Content{
							{Parts: []*genai.Part{genai.NewPartFromText(prompt.Value)}},
						}

						resp, err := genaiClient.Models.GenerateContent(context.Background(), model, content, config)
						if err != nil {
							return evaldo.MakeError(ps, fmt.Sprintf("Generation failed: %v", err))
						}

						responseText := resp.Text()
						if responseText == "" {
							return evaldo.MakeError(ps, "No response text returned from GenAI")
						}

						return *env.NewString(responseText)
					default:
						return evaldo.MakeError(ps, "Arg 3 must be a dictionary (options).")
					}
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a GenAI client.")
			}
		},
	},

	//
	// ##### Search with Google Grounding ##### "Functions for web search with Gemini grounding"
	//

	// Tests:
	// search-response: client .search "latest AI developments"
	// Args:
	// * client: GenAI client instance
	// * query: String - Search query
	// Returns:
	// * dict - Response containing text and sources
	"genai-client//Search": {
		Argsn: 2,
		Doc:   "Search Google and get intelligent summary with grounding. Returns both answer and sources.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch query := arg1.(type) {
				case env.String:
					genaiClient := client.Value.(*genai.Client)

					config := &genai.GenerateContentConfig{
						Tools: []*genai.Tool{
							{GoogleSearch: &genai.GoogleSearch{}},
						},
						Temperature: genai.Ptr(float32(1.0)),
					}

					content := []*genai.Content{
						{Parts: []*genai.Part{genai.NewPartFromText(query.Value)}},
					}

					resp, err := genaiClient.Models.GenerateContent(context.Background(), "gemini-2.5-flash", content, config)
					if err != nil {
						return evaldo.MakeError(ps, fmt.Sprintf("Search failed: %v", err))
					}

					result := make(map[string]any)
					result["text"] = *env.NewString(resp.Text())

					// Extract grounding sources
					sources := make([]env.Object, 0)
					if resp.Candidates != nil && len(resp.Candidates) > 0 {
						metadata := resp.Candidates[0].GroundingMetadata
						if metadata != nil && len(metadata.GroundingChunks) > 0 {
							seen := make(map[string]bool)
							for _, chunk := range metadata.GroundingChunks {
								if chunk.Web != nil && chunk.Web.URI != "" && !seen[chunk.Web.URI] {
									source := make(map[string]any)
									source["title"] = *env.NewString(chunk.Web.Title)
									source["url"] = *env.NewString(chunk.Web.URI)
									sources = append(sources, *env.NewDict(source))
									seen[chunk.Web.URI] = true
								}
							}
						}
					}
					result["sources"] = *env.NewBlock(*env.NewTSeries(sources))

					return *env.NewDict(result)
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (query).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a GenAI client.")
			}
		},
	},

	//
	// ##### Streaming Generation ##### "Functions for streaming text generation"
	//

	// Tests:
	// client .generate\stream "Tell a story" { |chunk| prn chunk }
	// Args:
	// * client: GenAI client instance
	// * prompt: String - Text prompt for generation
	// * callback: Block - Function to call for each chunk of streamed response
	// Returns:
	// * string - Complete response text after streaming is finished
	"genai-client//Generate\\stream": {
		Argsn: 3,
		Doc:   "Stream text generation with real-time chunks. Calls callback function for each piece of the response as it arrives.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch prompt := arg1.(type) {
				case env.String:
					switch callback := arg2.(type) {
					case env.Block:
						genaiClient := client.Value.(*genai.Client)

						content := []*genai.Content{
							{Parts: []*genai.Part{genai.NewPartFromText(prompt.Value)}},
						}

						streamIter := genaiClient.Models.GenerateContentStream(context.Background(), "gemini-2.5-flash", content, nil)

						var fullResponse strings.Builder
						for resp, err := range streamIter {
							if err != nil {
								return evaldo.MakeError(ps, fmt.Sprintf("Stream error: %v", err))
							}

							chunkText := resp.Text()
							if chunkText != "" {
								fullResponse.WriteString(chunkText)

								// Call the callback with the chunk
								ser := ps.Ser
								ps.Ser = callback.Series
								evaldo.EvalBlockInj(ps, *env.NewString(chunkText), true)
								ps.Ser = ser

								if ps.ErrorFlag {
									return ps.Res
								}
							}
						}

						return *env.NewString(fullResponse.String())
					default:
						return evaldo.MakeError(ps, "Arg 3 must be a block (callback function).")
					}
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a GenAI client.")
			}
		},
	},

	//
	// ##### Chat with Conversation History ##### "Functions for conversational AI"
	//

	// Tests:
	// conversation: [ { "role" "user" "content" "Hello!" } { "role" "model" "content" "Hi there!" } { "role" "user" "content" "How are you?" } ]
	// response: client .chat conversation
	// Args:
	// * client: GenAI client instance
	// * conversation: Block - Conversation format with role/content dicts
	// Returns:
	// * string - The AI's response text
	"genai-client//Chat": {
		Argsn: 2,
		Doc:   "Generate chat response using conversation history. Accepts conversation format with message history.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch conversation := arg1.(type) {
				case env.Block:
					genaiClient := client.Value.(*genai.Client)

					var contents []*genai.Content

					for i, item := range conversation.Series.S {
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
								contents = append(contents, &genai.Content{
									Role:  genai.RoleUser,
									Parts: []*genai.Part{genai.NewPartFromText(contentStr.Value)},
								})
							case "model":
								contents = append(contents, &genai.Content{
									Role:  genai.RoleModel,
									Parts: []*genai.Part{genai.NewPartFromText(contentStr.Value)},
								})
							default:
								return evaldo.MakeError(ps, fmt.Sprintf("Invalid role '%s' at index %d. Must be 'user' or 'model'", roleStr.Value, i))
							}
						} else {
							return evaldo.MakeError(ps, fmt.Sprintf("Message at index %d must be a dictionary", i))
						}
					}

					resp, err := genaiClient.Models.GenerateContent(context.Background(), "gemini-2.5-flash", contents, nil)
					if err != nil {
						return evaldo.MakeError(ps, fmt.Sprintf("Chat failed: %v", err))
					}

					responseText := resp.Text()
					if responseText == "" {
						return evaldo.MakeError(ps, "No response text returned from GenAI")
					}

					return *env.NewString(responseText)
				case env.String:
					// Simple string prompt
					genaiClient := client.Value.(*genai.Client)

					content := []*genai.Content{
						{Parts: []*genai.Part{genai.NewPartFromText(conversation.Value)}},
					}

					resp, err := genaiClient.Models.GenerateContent(context.Background(), "gemini-2.5-flash", content, nil)
					if err != nil {
						return evaldo.MakeError(ps, fmt.Sprintf("Chat failed: %v", err))
					}

					responseText := resp.Text()
					if responseText == "" {
						return evaldo.MakeError(ps, "No response text returned from GenAI")
					}

					return *env.NewString(responseText)
				default:
					return evaldo.MakeError(ps, "Arg 2 must be a string (prompt) or block (conversation).")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a GenAI client.")
			}
		},
	},

	//
	// ##### Model Management ##### "Functions for model information"
	//

	// Tests:
	// models: client .list-models
	// Args:
	// * client: GenAI client instance
	// Returns:
	// * block - List of available models
	"genai-client//List-models": {
		Argsn: 1,
		Doc:   "List all available Google GenAI models.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				genaiClient := client.Value.(*genai.Client)

				page, err := genaiClient.Models.List(context.Background(), nil)
				if err != nil {
					return evaldo.MakeError(ps, fmt.Sprintf("Failed to list models: %v", err))
				}

				var models []env.Object
				for _, model := range page.Items {
					modelDict := make(map[string]any)
					modelDict["name"] = *env.NewString(model.Name)
					modelDict["display-name"] = *env.NewString(model.DisplayName)
					modelDict["description"] = *env.NewString(model.Description)

					models = append(models, *env.NewDict(modelDict))
				}

				return *env.NewBlock(*env.NewTSeries(models))
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a GenAI client.")
			}
		},
	},

	//
	// ##### Terminal Search Utility ##### "Replicates the terminal search functionality from your example"
	//

	// Tests:
	// search-result: genai\search "latest AI developments"
	// Args:
	// * query: String - Search query
	// Returns:
	// * dict - Response containing text and sources, mimicking the terminal tool
	"search": {
		Argsn: 1,
		Doc:   "Terminal-style search function that replicates the behavior of your example script. Uses GEMINI_API_KEY environment variable.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch query := arg0.(type) {
			case env.String:
				ctx := context.Background()

				// Initialize the GenAI Client (automatically picks up GEMINI_API_KEY)
				client, err := genai.NewClient(ctx, nil)
				if err != nil {
					return evaldo.MakeError(ps, fmt.Sprintf("Failed to initialize GenAI client: %v", err))
				}

				// Configure the interaction with Search Grounding enabled
				config := &genai.GenerateContentConfig{
					Tools: []*genai.Tool{
						{GoogleSearch: &genai.GoogleSearch{}}, // Enlists Google Search
					},
					Temperature: genai.Ptr(float32(1.0)), // Recommended value for optimal grounding
				}

				content := []*genai.Content{
					{Parts: []*genai.Part{genai.NewPartFromText(query.Value)}},
				}

				// Use the model to generate content with the search tool
				resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", content, config)
				if err != nil {
					return evaldo.MakeError(ps, fmt.Sprintf("Search failed: %v", err))
				}

				result := make(map[string]any)
				result["answer"] = *env.NewString(resp.Text())

				// Extract and present the web sources used to ground the answer
				sources := make([]env.Object, 0)
				if resp.Candidates != nil && len(resp.Candidates) > 0 {
					metadata := resp.Candidates[0].GroundingMetadata
					if metadata != nil && len(metadata.GroundingChunks) > 0 {
						// Track duplicates so we don't spam with the same URL
						seen := make(map[string]bool)
						for _, chunk := range metadata.GroundingChunks {
							if chunk.Web != nil && chunk.Web.URI != "" && !seen[chunk.Web.URI] {
								source := make(map[string]any)
								source["title"] = *env.NewString(chunk.Web.Title)
								source["url"] = *env.NewString(chunk.Web.URI)
								sources = append(sources, *env.NewDict(source))
								seen[chunk.Web.URI] = true
							}
						}
					}
				}
				result["sources"] = *env.NewBlock(*env.NewTSeries(sources))

				return *env.NewDict(result)
			default:
				return evaldo.MakeError(ps, "Arg 1 must be a string (search query).")
			}
		},
	},
}
