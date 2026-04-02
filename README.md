# go-mistral

[![Go Reference](https://pkg.go.dev/badge/github.com/onkyou/go-mistral/mistral.svg)](https://pkg.go.dev/github.com/onkyou/go-mistral/mistral)
[![Go Report Card](https://goreportcard.com/badge/github.com/onkyou/go-mistral)](https://goreportcard.com/report/github.com/onkyou/go-mistral)
[![License](https://img.shields.io/github/license/thomas-marquis/mistral-client)](LICENSE)

An idiomatic, lightweight, and type-safe Go client for the [Mistral AI API](https://docs.mistral.ai/).

## Features

- **Chat Completions**: Synchronous and real-time streaming support.
- **Embeddings**: Create text embeddings for single or batch inputs.
- **Moderation & Classification**: Built-in support for Mistral's moderation and classification models.
- **Modern Go**: Uses Go 1.23+ iterators (`iter.Seq2`) for powerful and clean streaming.
- **Type-Safe**: Deeply modeled requests and responses with full validation logic.
- **Functional Helpers**: Convenient constructors for common message types.

## Installation

```bash
go get github.com/onkyou/go-mistral/mistral
```

## Quick Start

### Initialize Client

```go
import "github.com/onkyou/go-mistral/mistral"

client, err := mistral.NewClient(&mistral.ClientConfig{
    APIKey: "your-api-key",
})
```

### Chat Completion

```go
resp, _, err := client.Chat.Complete(ctx, &mistral.ChatCompletionRequest{
    Model: mistral.ModelMistralLargeLatest,
    Messages: []*mistral.ChatMessage{
        mistral.NewUserMessage("Moin! How's the weather in Hamburg?"),
    },
})
if err == nil {
    fmt.Println(resp.Choices[0].Message.Content)
}
```

### Streaming Chat

```go
req := &mistral.ChatCompletionRequest{
    Model: mistral.ModelMistralLargeLatest,
    Messages: []*mistral.ChatMessage{
        mistral.NewUserMessage("Write a short poem about Go."),
    },
}

for chunk, err := range client.Chat.Stream(ctx, req) {
    if err != nil {
        log.Fatal(err)
    }
    if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
        fmt.Print(chunk.Choices[0].Delta.Content)
    }
}
```

### Embeddings

```go
resp, _, err := client.Embedding.Create(ctx, &mistral.EmbeddingRequest{
    Model: mistral.ModelMistralEmbed,
    Input: []string{"Embed this sentence."},
})
if err == nil {
    fmt.Printf("Vector: %v\n", resp.Data[0].Embedding)
}
```

### Moderation

```go
resp, _, err := client.Classifiers.Moderate(ctx, &mistral.ModerateRequest{
    Model: mistral.ModelMistralModerationLatest,
    Input: "I want to say something mean.",
})
if err == nil {
    fmt.Printf("Is flagged: %v\n", resp.Results[0].Categories)
}
```

## Examples

Check the [examples/](https://github.com/onkyou/go-mistral/tree/main/examples) directory for complete, runnable examples:

- [Chat Completion](examples/chat/main.go)
- [Streaming Chat](examples/chat_stream/main.go)
- [Embeddings](examples/embedding/main.go)
- [Tool Call](examples/tool_call/main.go)

## Configuration

The client can be customized via `ClientConfig`:

```go
cfg := &mistral.ClientConfig{
    APIKey:     "your-api-key",
    BaseURL:    "https://api.mistral.ai/",
    HTTPClient: &http.Client{Timeout: 30 * time.Second},
}
client, err := mistral.NewClient(cfg)
```

## License

MIT
