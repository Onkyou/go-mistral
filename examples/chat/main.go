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

	client, err := mistral.NewClient(mistral.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	messages := []mistral.ChatMessage{
		{Role: mistral.RoleUser, Content: "Ahoy! Do seals prefer talking to whales or dolphins?"},
	}
	model := mistral.ModelMistralLargeLatest

	fmt.Println("Requesting completion...")
	resp, _, err := client.Chat.Complete(ctx, model, messages, mistral.WithTemperature(0.7))
	if err != nil {
		log.Fatalf("Failed to complete chat: %v", err)
	}

	fmt.Printf("Assistant: %s\n", resp.Choices[0].Message.Content)
}
