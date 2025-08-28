# Travel Agent Example

This example demonstrates a travel agent that can book flights and hotels using two mock tools. The agent receives a travel request and automatically uses the appropriate tools to complete the booking.

## 🎯 What It Does

The travel agent:
1. **Receives a travel request** from Copenhagen to Barcelona
2. **Uses the `book_flight` tool** to book a flight
3. **Uses the `book_hotel` tool** to book a hotel for 4 nights
4. **Coordinates the entire process** automatically

## 🛠️ Tools Available

### 1. Flight Booking Tool (`book_flight`)
- **Purpose**: Books flights between cities
- **Input**: Departure city, arrival city, date, class, passengers
- **Output**: Flight details including flight number, times, airline, price
- **Mock Data**: Returns Scandinavian Airlines flight SK1234 at €299.99

### 2. Hotel Booking Tool (`book_hotel`)
- **Purpose**: Books hotels in specific cities
- **Input**: City, check-in/check-out dates, room type, guests
- **Output**: Hotel details including name, address, price per night
- **Mock Data**: Returns Hotel Barcelona Palace at €189.99/night

## 🏗️ Architecture

The example demonstrates:
- **Agent Pattern**: An agent that orchestrates tool usage
- **Tool Integration**: How tools are integrated with the LLM
- **Schema Generation**: Automatic JSON schema generation from Go structs
- **Mock Tools**: Realistic tool implementations that return fake data
- **Conversation Flow**: How the agent handles multi-step tasks

## 🚀 Running the Example

### Prerequisites
1. Set your OpenAI API key:
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

2. Make sure you have the required dependencies:
   ```bash
   go mod tidy
   ```

### Run the Example
```bash
cd examples/travel_agent
go run .
```

## 📋 Expected Output

The agent will:
1. Start with the travel request
2. Automatically call the flight booking tool
3. Automatically call the hotel booking tool
4. Provide a summary of the bookings
5. Show the complete conversation flow

## 🔧 Key Components

### Data Structures
- `Flight`: Represents a flight booking with all details
- `Hotel`: Represents a hotel booking with all details
- `FlightBookingRequest`: Input schema for flight booking
- `HotelBookingRequest`: Input schema for hotel booking

### Tools
- `FlightBookingTool`: Implements the flight booking logic
- `HotelBookingTool`: Implements the hotel booking logic

### Agent
- Uses the OpenAI LLM to understand the request
- Automatically selects and calls the appropriate tools
- Manages the conversation flow

## 🎨 Customization

You can easily customize this example by:

1. **Adding more tools**: Create new tools for car rentals, restaurant reservations, etc.
2. **Modifying mock data**: Change the fake flight/hotel details
3. **Adding validation**: Implement real validation logic
4. **Integrating real APIs**: Replace mock tools with real booking APIs

## 🔍 Learning Points

This example demonstrates:
- How to create tools that implement the `llm.Tool` interface
- How to use the schema generation package for OpenAI compatibility
- How agents automatically orchestrate multi-tool workflows
- How to handle complex, multi-step user requests
- Best practices for tool design and implementation

## 🚨 Important Notes

- **Mock Data**: This example uses fake data for demonstration purposes
- **API Key Required**: You need a valid OpenAI API key to run this example
- **Tool Schemas**: All tools use automatically generated JSON schemas for OpenAI compatibility
- **Error Handling**: The example includes basic error handling for robustness 