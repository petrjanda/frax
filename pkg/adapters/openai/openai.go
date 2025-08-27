package openai

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"

	"github.com/petrjanda/frax/pkg/llm"
)

// OpenAIAdapter implements the LLM interface using OpenAI's API
type OpenAIAdapter struct {
	client *openai.Client
	model  string
}

// OpenAIAdapterOpts represents options for configuring the OpenAI adapter
type OpenAIAdapterOpts = func(*OpenAIAdapter)

// WithModel sets the model to use for the OpenAI adapter
func WithModel(model string) OpenAIAdapterOpts {
	return func(a *OpenAIAdapter) {
		a.model = model
	}
}

// NewOpenAIAdapter creates a new OpenAI adapter with the given API key and options
func NewOpenAIAdapter(apiKey string, opts ...OpenAIAdapterOpts) (*OpenAIAdapter, error) {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	adapter := &OpenAIAdapter{
		client: &client,
		model:  "gpt-4o", // default model
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter, nil
}

// Invoke implements the LLM interface by calling OpenAI's API
func (a *OpenAIAdapter) Invoke(ctx context.Context, request *llm.LLMRequest) (*llm.LLMResponse, error) {
	// Convert our messages to OpenAI format
	messages := a.convertMessages(request.History)

	// Prepare the chat completion request
	chatReq := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(a.model),
		Messages: messages,
	}

	// Add tools if provided and tool usage is enabled
	if request.ToolUsage && len(request.Tools) > 0 {
		tools := a.convertTools(request.Tools)
		chatReq.Tools = tools
		// For now, we'll omit ToolChoice to let OpenAI use the default behavior
		// This should resolve the "Missing required parameter: 'tool_choice.function'" error

		// Debug: Log what we're sending
		fmt.Printf("🔧 Debug: Sending %d tools to OpenAI\n", len(tools))
		for i, tool := range tools {
			fmt.Printf("  Tool %d: %+v\n", i+1, tool)
		}
	}

	// Debug: Log the request we're about to send
	fmt.Printf("📤 Debug: Sending request to OpenAI:\n")
	fmt.Printf("  Model: %s\n", chatReq.Model)
	fmt.Printf("  Messages: %d\n", len(chatReq.Messages))
	fmt.Printf("  Tools: %d\n", len(chatReq.Tools))
	// Note: ToolChoice is not set to avoid the "Missing required parameter" error

	// Make the API call
	resp, err := a.client.Chat.Completions.New(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// Convert the response to our format
	response := llm.NewLLMResponse()

	// Add the assistant's message
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		if choice.Message.Content != "" {
			// Create a text message from the content
			textMsg := &llm.UserMessage{Content: choice.Message.Content}
			response.AddMessage(textMsg)
		}

		// Handle tool calls if any
		if choice.Message.ToolCalls != nil {
			for _, toolCall := range choice.Message.ToolCalls {
				// Convert OpenAI tool call to our format
				ourToolCall := &llm.ToolCall{
					Name: toolCall.Function.Name,
					Args: json.RawMessage(toolCall.Function.Arguments),
				}
				response.AddToolCall(ourToolCall)
			}
		}
	}

	return response, nil
}

// convertMessages converts our Message interface to OpenAI's format
func (a *OpenAIAdapter) convertMessages(messages []llm.Message) []openai.ChatCompletionMessageParamUnion {
	var openaiMessages []openai.ChatCompletionMessageParamUnion

	for _, msg := range messages {
		switch m := msg.(type) {
		case *llm.UserMessage:
			openaiMessages = append(openaiMessages, openai.UserMessage(m.Content))
		case *llm.ToolResultMessage:
			// Convert tool result to tool message
			openaiMessages = append(openaiMessages, openai.ToolMessage(string(m.Result), m.ToolCall.Name))
		}
	}

	return openaiMessages
}

// convertTools converts our Tool interface to OpenAI's format
func (a *OpenAIAdapter) convertTools(tools []llm.Tool) []openai.ChatCompletionToolUnionParam {
	var openaiTools []openai.ChatCompletionToolUnionParam

	for _, tool := range tools {
		// Parse the JSON schema to convert to FunctionParameters
		var params map[string]any
		if err := json.Unmarshal(tool.InputSchemaRaw(), &params); err != nil {
			// If we can't parse the schema, use an empty object
			params = make(map[string]any)
		}

		functionDef := shared.FunctionDefinitionParam{
			Name:        tool.Name(),
			Description: openai.String(tool.Description()),
			Parameters:  shared.FunctionParameters(params),
		}

		openaiTool := openai.ChatCompletionFunctionTool(functionDef)
		openaiTools = append(openaiTools, openaiTool)
	}

	return openaiTools
}
