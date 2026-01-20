# Ollama Integration for Rye

This module provides Rye bindings for the [Ollama](https://ollama.ai/) API, enabling local LLM inference, embeddings, and model management directly from Rye.

## Building with Ollama Support

To build Rye with Ollama support, use the `b_ollama` build tag:

```bash
go build -tags "b_ollama"
```

## Requirements

- Ollama must be installed and running locally (default: `http://localhost:11434`)
- Set `OLLAMA_HOST` environment variable if using a different host

## Usage Examples

### Creating a Client

```rye
; Create client from environment (uses OLLAMA_HOST or defaults to localhost:11434)
client: ollama
```

### Embeddings

The primary use case - creating vector embeddings for text:

```rye
; Create a single embedding
client: ollama
embedding: client .embed "bge-m3" "Hello, world!"
; embedding is a vector of floats

; Create multiple embeddings
embeddings: client .embed\many "bge-m3" { "Hello" "World" "Test" }
; embeddings is a block of vectors
```

### Chat Completions

```rye
; Simple chat
client: ollama
response: client .chat "llama2" "What is the capital of France?"
print response

; Chat with conversation history
messages: [
    dict { "role" "system" "content" "You are a helpful assistant." }
    dict { "role" "user" "content" "Hello!" }
]
response: client .chat\messages "llama2" messages

; Streaming chat (real-time response)
client .chat\stream "llama2" "Tell me a story" { .print }
```

### Text Generation

```rye
; Simple text completion
response: client .generate "llama2" "The sky is"
```

### Model Management

```rye
; List available models
models: client .list-models
for models { -> "name" |print }

; Show model details
info: client .show-model "llama2"
print info -> "parameters"

; Pull a new model
client .pull-model "mistral"
```

## Available Functions

| Function | Description |
|----------|-------------|
| `ollama` | Create client from environment |
| `ollama\url` | Create client with specific URL |
| `.embed` | Create embedding for text |
| `.embed\many` | Create embeddings for multiple texts |
| `.chat` | Simple chat completion |
| `.chat\messages` | Chat with conversation history |
| `.chat\stream` | Streaming chat completion |
| `.generate` | Text generation |
| `.list-models` | List available models |
| `.show-model` | Show model details |
| `.pull-model` | Download a model |

## Embedding Models

Popular embedding models for use with `.embed`:

- `bge-m3` - Multilingual embeddings
- `nomic-embed-text` - General purpose embeddings  
- `mxbai-embed-large` - High quality embeddings
- `all-minilm` - Fast, lightweight embeddings

## Chat/Generation Models

Popular models for use with `.chat` and `.generate`:

- `llama2` - Meta's Llama 2
- `mistral` - Mistral AI
- `codellama` - Code-focused Llama
- `phi` - Microsoft's Phi model
- `neural-chat` - Intel's chat model

## Environment Variables

- `OLLAMA_HOST` - Ollama server URL (default: `http://localhost:11434`)
