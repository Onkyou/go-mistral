# go-mistral

[![Go Reference](https://pkg.go.dev/badge/github.com/onkyou/go-mistral/mistral.svg)](https://pkg.go.dev/github.com/onkyou/go-mistral/mistral)
[![Go Report Card](https://goreportcard.com/badge/github.com/onkyou/go-mistral)](https://goreportcard.com/report/github.com/onkyou/go-mistral)

An idiomatic, lightweight, and type-safe Go client for the [Mistral AI API](https://docs.mistral.ai/).

## Features

- **Chat Completions**: Synchronous and real-time streaming support.
- **Embeddings**: Create text embeddings for single or batch inputs.
- **Functional Options**: Clean and flexible configuration using the options pattern.
- **SSE Streaming**: Idiomatic Go channels for streaming responses.

## Installation

```bash
go get github.com/onkyou/go-mistral/mistral
```

## Quick Start

### Initialize Client

```go
import "github.com/onkyou/go-mistral/mistral"

client, err := mistral.NewClient(mistral.WithAPIKey("your-api-key"))
```

### Chat Completion

```go
messages := []mistral.ChatMessage{
    {Role: mistral.RoleUser, Content: "Moin! How's the weather in Hamburg?"},
}

resp, _, err := client.Chat.Complete(ctx, mistral.ModelMistralLargeLatest, messages)
if err == nil {
    fmt.Println(resp.Choices[0].Message.Content)
}
```

### Streaming Chat

```go
dataChan, errChan := client.Chat.Stream(ctx, model, messages)

for chunk := range dataChan {
    if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
        fmt.Print(chunk.Choices[0].Delta.Content)
    }
}

if err := <-errChan; err != nil {
    log.Fatal(err)
}
```

### Embeddings

```go
resp, _, err := client.Embedding.Create(ctx, mistral.ModelMistralEmbed, "Embed this sentence.")
if err == nil {
    fmt.Printf("Vector: %v\n", resp.Data[0].Embedding)
}
```

## Examples

Check the [examples/](https://github.com/onkyou/go-mistral/tree/main/examples) directory for complete, runnable examples of:

- [Chat Completion](examples/chat/main.go)
- [Streaming Chat](examples/chat_stream/main.go)
- [Embeddings](examples/embedding/main.go)

## License

MIT
