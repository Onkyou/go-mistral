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
		{Role: mistral.RoleUser, Content: "Moin! How would a seagull muse about Hamburg?"},
	}
	model := mistral.ModelMistralLargeLatest

	dataChan, errChan := client.Chat.Stream(ctx, model, messages, mistral.WithTemperature(0.7))

	fmt.Print("Assistant: ")
	for chunk := range dataChan {
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}

	if err := <-errChan; err != nil {
		log.Fatalf("\nError in stream: %v", err)
	}
	fmt.Println()
}
