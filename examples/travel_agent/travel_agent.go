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
	Departure    time.Time `json:"departure" jsonschema:"required,description=Must be in RFC3339 format (e.g. 2024-01-01T15:04:05Z)"`
	Arrival      time.Time `json:"arrival" jsonschema:"required,description=Must be in RFC3339 format (e.g. 2024-01-01T15:04:05Z)"`
	Airline      string    `json:"airline" jsonschema:"required"`
	Price        float64   `json:"price" jsonschema:"required"`
	Class        string    `json:"class" jsonschema:"required"`
}

// Hotel represents a hotel booking
type Hotel struct {
	HotelName     string    `json:"hotel_name" jsonschema:"required"`
	City          string    `json:"city" jsonschema:"required"`
	CheckIn       time.Time `json:"check_in" jsonschema:"required,description=Must be in RFC3339 format (e.g. 2024-01-01T15:04:05Z)"`
	CheckOut      time.Time `json:"check_out" jsonschema:"required,description=Must be in RFC3339 format (e.g. 2024-01-01T15:04:05Z)"`
	RoomType      string    `json:"room_type" jsonschema:"required"`
	PricePerNight float64   `json:"price_per_night" jsonschema:"required"`
	TotalPrice    float64   `json:"total_price" jsonschema:"required"`
	Address       string    `json:"address" jsonschema:"required"`
}

// FlightBookingRequest represents the input for booking a flight
type FlightBookingRequest struct {
	From       string    `json:"from" jsonschema:"required"`
	To         string    `json:"to" jsonschema:"required"`
	Date       time.Time `json:"date" jsonschema:"required,description=Must be in RFC3339 format (e.g. 2024-01-01T15:04:05Z)"`
	Class      string    `json:"class" jsonschema:"required"`
	Passengers int       `json:"passengers" jsonschema:"required"`
}

// HotelBookingRequest represents the input for booking a hotel
type HotelBookingRequest struct {
	City     string    `json:"city" jsonschema:"required"`
	CheckIn  time.Time `json:"check_in" jsonschema:"required,description=Must be in RFC3339 format (e.g. 2024-01-01T15:04:05Z)"`
	CheckOut time.Time `json:"check_out" jsonschema:"required,description=Must be in RFC3339 format (e.g. 2024-01-01T15:04:05Z)"`
	RoomType string    `json:"room_type" jsonschema:"required"`
	Guests   int       `json:"guests" jsonschema:"required"`
}

type TravelItinerary struct {
	Flight Flight `json:"flight"`
	Hotel  Hotel  `json:"hotel"`
}

// Tool functions - clean and typed!
func bookFlight(ctx context.Context, request FlightBookingRequest) (Flight, error) {
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

	return flight, nil
}

var flightTool = llm.CreateTool(
	"book_flight",
	`Book a flight from one city to another. 
	Provide departure and arrival cities, preferred dates, and travel class.`,
	bookFlight,
)

func bookHotel(ctx context.Context, request HotelBookingRequest) (Hotel, error) {
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

	return hotel, nil
}

var hotelTool = llm.CreateTool(
	"book_hotel",
	`Book a hotel in a specific city. 
	Provide city, check-in and check-out dates, and room type preferences.`,
	bookHotel,
)

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

	// Create agent with tools and retry configuration
	agent := llm.NewAgent(openaiLLM,
		llm.NewToolbox(flightTool, hotelTool),

		llm.WithMaxRetries(3),                    // Allow up to 3 retries
		llm.WithRetryDelay(200*time.Millisecond), // Start with 200ms delay
		llm.WithRetryBackoff(1.5),                // 1.5x backoff multiplier
		llm.WithOutputSchema(
			schemas.
				NewOpenAISchemaGenerator().
				MustGenerateSchema(TravelItinerary{}),
		),
	)

	// Create conversation history with the travel request
	history := llm.NewHistory(
		llm.NewUserMessage(`
			I need to book travel to Barcelona. 
			I will be flying from Copenhagen. 
			Please book me a flight and a hotel for a 4-night stay. 
			I prefer premium economy class for the flight and a luxury room for the hotel. 
			I will flying out at 1st Nov.
		`),
	)

	// Run the agent
	response, err := agent.Invoke(ctx, llm.NewLLMRequest(history,
		llm.WithSystem(`
		You are a travel agent. You are given a task to book travel to a specific city.
		You have two tools available to you: book_flight and book_hotel.
		This is required for both flight and hotel bookings.

		When working with dates, always use RFC3339 format (e.g. 2024-01-01T15:04:05Z) for all date/time values. 
		Example: "2024-01-01T15:04:05Z" for January 1st, 2024 at 3pm.
	`),
		llm.WithTemperature(0.0),
		llm.WithMaxCompletionTokens(1000),
	))
	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	// Print the conversation
	for _, msg := range response.Messages {
		payload, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("Failed to marshal message: %v", err)
		}
		fmt.Println(string(payload))
	}
}
