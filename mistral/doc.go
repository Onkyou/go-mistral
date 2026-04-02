// Package mistral is an inofficial Go client for the Mistral AI API.
// refer to https://docs.mistral.ai/api
//
// It provides a clean, idiomatic Go interface for interacting with Mistral's most powerful models,
// including support for Chat Completions (with tool calls and streaming), Embeddings,
// and the new Moderation and Classification endpoints.
//
// Usage:
//
//	client, err := mistral.NewClient(&mistral.ClientConfig{APIKey: "your-api-key"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	resp, _, err := client.Chat.Complete(ctx, &mistral.ChatCompletionRequest{
//	    Model: mistral.ModelMistralLargeLatest,
//	    Messages: []*mistral.ChatMessage{
//	        mistral.NewUserMessage("Hello!"),
//	    },
//	})
package mistral
