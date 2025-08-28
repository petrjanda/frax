package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/petrjanda/frax/pkg/adapters/openai"
	"github.com/petrjanda/frax/pkg/adapters/openai/schemas"
	"github.com/petrjanda/frax/pkg/llm"
)

// Flight represents a flight booking
type Flight struct {
	FlightNumber string    `json:"flight_number" jsonschema:"required"`
	From         string    `json:"from" jsonschema:"required"`
	To           string    `json:"to" jsonschema:"required"`
	Departure    time.Time `json:"departure" jsonschema:"required"`
	Arrival      time.Time `json:"arrival" jsonschema:"required"`
	Airline      string    `json:"airline" jsonschema:"required"`
	Price        float64   `json:"price" jsonschema:"required"`
	Class        string    `json:"class" jsonschema:"required"`
}

// Hotel represents a hotel booking
type Hotel struct {
	HotelName     string    `json:"hotel_name" jsonschema:"required"`
	City          string    `json:"city" jsonschema:"required"`
	CheckIn       time.Time `json:"check_in" jsonschema:"required"`
	CheckOut      time.Time `json:"check_out" jsonschema:"required"`
	RoomType      string    `json:"room_type" jsonschema:"required"`
	PricePerNight float64   `json:"price_per_night" jsonschema:"required"`
	TotalPrice    float64   `json:"total_price" jsonschema:"required"`
	Address       string    `json:"address" jsonschema:"required"`
}

// FlightBookingTool is a mock tool for booking flights
type FlightBookingTool struct{}

func (f *FlightBookingTool) Name() string {
	return "book_flight"
}

func (f *FlightBookingTool) Description() string {
	return "Book a flight from one city to another. Provide departure and arrival cities, preferred dates, and travel class."
}

func (f *FlightBookingTool) InputSchemaRaw() json.RawMessage {
	generator := schemas.NewOpenAISchemaGenerator()
	return generator.MustGenerateSchema(FlightBookingRequest{})
}

func (f *FlightBookingTool) OutputSchemaRaw() json.RawMessage {
	generator := schemas.NewOpenAISchemaGenerator()
	return generator.MustGenerateSchema(Flight{})
}

func (f *FlightBookingTool) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var request FlightBookingRequest
	if err := json.Unmarshal(args, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Mock flight booking - generate fake flight details
	departure := time.Now().AddDate(0, 0, 7) // 1 week from now
	arrival := departure.Add(3 * time.Hour)  // 3 hour flight

	flight := Flight{
		FlightNumber: "SK1234",
		From:         request.From,
		To:           request.To,
		Departure:    departure,
		Arrival:      arrival,
		Airline:      "Scandinavian Airlines",
		Price:        299.99,
		Class:        request.Class,
	}

	result, err := json.Marshal(flight)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flight: %w", err)
	}

	return json.RawMessage(result), nil
}

// HotelBookingTool is a mock tool for booking hotels
type HotelBookingTool struct{}

func (h *HotelBookingTool) Name() string {
	return "book_hotel"
}

func (h *HotelBookingTool) Description() string {
	return "Book a hotel in a specific city. Provide city, check-in and check-out dates, and room type preferences."
}

func (h *HotelBookingTool) InputSchemaRaw() json.RawMessage {
	generator := schemas.NewOpenAISchemaGenerator()
	return generator.MustGenerateSchema(HotelBookingRequest{})
}

func (h *HotelBookingTool) OutputSchemaRaw() json.RawMessage {
	generator := schemas.NewOpenAISchemaGenerator()
	return generator.MustGenerateSchema(Hotel{})
}

func (h *HotelBookingTool) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var request HotelBookingRequest
	if err := json.Unmarshal(args, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Mock hotel booking - generate fake hotel details
	checkIn := time.Now().AddDate(0, 0, 7) // 1 week from now
	checkOut := checkIn.AddDate(0, 0, 4)   // 4 nights stay

	hotel := Hotel{
		HotelName:     "Hotel Barcelona Palace",
		City:          request.City,
		CheckIn:       checkIn,
		CheckOut:      checkOut,
		RoomType:      request.RoomType,
		PricePerNight: 189.99,
		TotalPrice:    189.99 * 4,
		Address:       "Carrer de Balmes, 142, 08008 Barcelona, Spain",
	}

	result, err := json.Marshal(hotel)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hotel: %w", err)
	}

	return json.RawMessage(result), nil
}

// FlightBookingRequest represents the input for booking a flight
type FlightBookingRequest struct {
	From       string    `json:"from" jsonschema:"required"`
	To         string    `json:"to" jsonschema:"required"`
	Date       time.Time `json:"date" jsonschema:"required"`
	Class      string    `json:"class" jsonschema:"required"`
	Passengers int       `json:"passengers" jsonschema:"required"`
}

// HotelBookingRequest represents the input for booking a hotel
type HotelBookingRequest struct {
	City     string    `json:"city" jsonschema:"required"`
	CheckIn  time.Time `json:"check_in" jsonschema:"required"`
	CheckOut time.Time `json:"check_out" jsonschema:"required"`
	RoomType string    `json:"room_type" jsonschema:"required"`
	Guests   int       `json:"guests" jsonschema:"required"`
}

func main() {
	ctx := context.Background()

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI adapter
	openaiLLM, err := openai.NewOpenAIAdapter(apiKey, openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	// Create travel booking tools
	tools := []llm.Tool{
		&FlightBookingTool{},
		&HotelBookingTool{},
	}

	// Create agent with tools and retry configuration
	agent := llm.NewAgent(openaiLLM, tools,
		llm.WithMaxRetries(3),                    // Allow up to 3 retries
		llm.WithRetryDelay(200*time.Millisecond), // Start with 200ms delay
		llm.WithRetryBackoff(1.5),                // 1.5x backoff multiplier
	)

	// Create conversation history with the travel request
	history := llm.NewHistory(
		llm.NewUserMessage("I need to book travel to Barcelona. I will be flying from Copenhagen. Please book me a flight and a hotel for a 4-night stay starting next week. I prefer economy class for the flight and a standard room for the hotel. I will travel 1st Nov out and 3rd Nov back."),
	)

	fmt.Println("🚀 Starting Travel Agent...")
	fmt.Println("📍 Task: Book travel to Barcelona from Copenhagen")
	fmt.Println("🛫 Tools available: book_flight, book_hotel")
	fmt.Println("============================================================")

	// Create a request with the travel task
	request := llm.NewLLMRequest(history)

	// Run the agent
	response, err := agent.Invoke(ctx, request)
	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	fmt.Println("\n============================================================")
	fmt.Println("🎉 Travel Agent completed the task!")
	fmt.Println("📝 Final conversation:")
	fmt.Println("============================================================")

	// Print the conversation
	for i, msg := range response.Messages {
		switch m := msg.(type) {
		case *llm.UserMessage:
			fmt.Printf("\n👤 User %d: %s\n", i+1, m.Content)
		case *llm.ToolResultMessage:
			fmt.Printf("\n🔧 Tool Result %d: %s\n", i+1, string(m.Result))
		case *llm.AssistantMessage:
			fmt.Printf("\n🤖 Assistant %d: %s\n", i+1, m.Content)
		default:
			fmt.Printf("\n💬 Message %d: %+v\n", i+1, msg)
		}
	}

	fmt.Println("\n============================================================")
	fmt.Println("✨ Travel booking completed successfully!")
}
