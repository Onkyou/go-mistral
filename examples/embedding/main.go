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
	input := "Embed this sentence."
	model := mistral.ModelMistralEmbed

	resp, _, err := client.Embedding.Create(ctx, model, []string{input})
	if err != nil {
		log.Fatalf("Failed to create embedding: %v", err)
	}

	fmt.Printf("Embedding for %q:\n", input)
	if len(resp.Data) > 0 {
		fmt.Printf("Vector (len=%d), showing first 5 elements: %v...\n", len(resp.Data[0].Embedding), resp.Data[0].Embedding[:5])
	}
}
