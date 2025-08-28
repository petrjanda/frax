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
	history := append(llm.NewHistory(llm.NewSystemMessage(request.System)), request.History...)

	// Prepare the chat completion request
	chatReq := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(a.model),
		Messages: a.convertMessages(history),
	}

	// Handle tool usage based on the ToolUsage strategy
	if request.ToolUsage != nil && len(request.Tools) > 0 {
		tools := a.convertTools(request.Tools)
		chatReq.Tools = tools

		// Convert tool usage to OpenAI format
		toolChoice, err := convertToolUsage(request.ToolUsage, request.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool usage: %w", err)
		}

		if toolChoice != nil {
			chatReq.ToolChoice = *toolChoice
		}
	}

	// Pretty print the chat request for debugging
	// reqJSON, err := json.MarshalIndent(chatReq, "", "  ")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	// }
	// fmt.Println(string(reqJSON))

	resp, err := a.client.Chat.Completions.New(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	response := llm.NewLLMResponse()

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		if choice.Message.Content != "" {
			textMsg := &llm.AssistantMessage{Content: choice.Message.Content}
			response.AddMessage(textMsg)
		}

		if choice.Message.ToolCalls != nil {
			for _, toolCall := range choice.Message.ToolCalls {
				ourToolCall := &llm.ToolCall{
					ID:   toolCall.ID,
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
		case *llm.AssistantMessage:
			openaiMessages = append(openaiMessages, openai.AssistantMessage(m.Content))
		case *llm.SystemMessage:
			openaiMessages = append(openaiMessages, openai.SystemMessage(m.Content))

		case *llm.ToolCallMessage:
			// Convert tool call to assistant message with tool_calls
			asst := openai.ChatCompletionAssistantMessageParam{
				Role: "assistant",
				ToolCalls: []openai.ChatCompletionMessageToolCallUnionParam{
					{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: m.ToolCall.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      m.ToolCall.Name,
								Arguments: string(m.ToolCall.Args),
							},
							Type: "function",
						},
					},
				},
			}
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessageParamUnion{OfAssistant: &asst})

		case *llm.ToolResultMessage:
			// Convert tool result to tool message
			openaiMessages = append(openaiMessages, openai.ToolMessage(string(m.Result), m.ToolCall.ID))

		case *llm.ToolErrorMessage:
			// Don't send tool error messages directly to OpenAI
			// Instead, we'll handle retries differently to avoid API violations
			// This case should not occur in normal operation with the new retry mechanism
			continue
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
