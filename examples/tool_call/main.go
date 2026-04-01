package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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

	// 1. Define the tool parameters using a typed struct
	type RandomNumberArgs struct {
		Min int `json:"min" mistral:"The minimum value (inclusive)."`
		Max int `json:"max" mistral:"The maximum value (inclusive)."`
	}

	// Define the list of tools available to the model
	tools := []mistral.Tool{
		mistral.NewFunctionTool(
			"get_random_number",
			"Generates a random number between a minimum and maximum value.",
			RandomNumberArgs{},
		),
	}

	// 2. Initial request to the model
	messages := []mistral.ChatMessage{
		mistral.NewUserMessage("Generate a random number between 1 and 100, and then write a four-line haiku about that number."),
	}

	model := mistral.ModelMistralLargeLatest

	fmt.Println("Requesting completion with tool...")
	resp, _, err := client.Chat.Complete(
		ctx,
		model,
		messages,
		mistral.WithTools(tools),
		mistral.WithToolChoice(mistral.ToolChoiceAuto),
	)
	if err != nil {
		log.Fatalf("Failed to complete chat: %v", err)
	}

	assistantMsg := resp.Choices[0].Message
	messages = append(messages, *assistantMsg)

	// 3. Check for tool calls
	if len(assistantMsg.ToolCalls) > 0 {
		for _, tc := range assistantMsg.ToolCalls {
			if tc.Function.Name == "get_random_number" {
				fmt.Printf("Model is calling tool: %s with arguments: %v\n", tc.Function.Name, tc.Function.Arguments)

				args, err := mistral.UnmarshalArguments[RandomNumberArgs](tc)
				if err != nil {
					log.Fatalf("Failed to unmarshal arguments: %v", err)
				}

				// Execute local function
				result := rand.Intn(args.Max-args.Min+1) + args.Min
				fmt.Printf("Local function result: %d\n", result)

				// Add tool response to messages using the new functional options
				messages = append(messages, mistral.NewToolMessage(fmt.Sprintf("%d", result), mistral.WithToolCallID(tc.ID)))
			}
		}

		// 4. Send the tool result back to the model for the final haiku
		fmt.Println("Sending tool result back to model...")
		resp, _, err = client.Chat.Complete(ctx, model, messages)
		if err != nil {
			log.Fatalf("Failed to get final response: %v", err)
		}

		fmt.Printf("\nAssistant Result:\n%s\n", resp.Choices[0].Message.Content)
	} else {
		fmt.Printf("Assistant: %s\n", assistantMsg.Content)
	}
}
