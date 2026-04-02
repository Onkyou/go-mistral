package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/onkyou/go-mistral/mistral"
)

func main() {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		log.Fatal("MISTRAL_API_KEY environment variable is not set")
	}

	client, err := mistral.NewClient(&mistral.ClientConfig{APIKey: apiKey})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	messages := []*mistral.ChatMessage{
		{Role: mistral.ChatMessageRoleUser, Content: "Moin! How would a seagull muse about Hamburg?"},
	}
	model := mistral.ModelMistralLargeLatest

	fmt.Print("Assistant: ")
	req := &mistral.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: func() *float64 { f := 0.7; return &f }(),
	}
	for chunk, err := range client.Chat.Stream(ctx, req) {
		if err != nil {
			log.Fatalf("\nError in stream: %v", err)
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}
	fmt.Println()
}
