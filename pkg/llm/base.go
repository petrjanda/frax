package llm

import (
	"context"
)

// LLM represents a language model that can process requests and generate responses
type LLM interface {
	Invoke(ctx context.Context, request *LLMRequest) (*LLMResponse, error)
}

// LLMRequest represents a request to the LLM
type LLMRequest struct {
	System    string
	History   []Message
	Tools     []Tool
	ToolUsage ToolUsage
}

// LLMRequestOpts represents options for configuring an LLM request
type LLMRequestOpts = func(*LLMRequest)

// WithToolUsage sets the tool usage strategy for the request
func WithToolUsage(toolUsage ToolUsage) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.ToolUsage = toolUsage
	}
}

// WithTools adds tools to the request
func WithTools(tools ...Tool) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.Tools = append(r.Tools, tools...)
	}
}

// WithSystem sets the system message for the request
func WithSystem(system string) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.System = system
	}
}

// NewLLMRequest creates a new LLM request with the given options
func NewLLMRequest(history History, opts ...LLMRequestOpts) *LLMRequest {
	r := &LLMRequest{
		History:   history,
		ToolUsage: AutoToolSelection(), // Default to auto tool selection
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *LLMRequest) Clone(opts ...LLMRequestOpts) *LLMRequest {
	req := &LLMRequest{
		History:   r.History,
		ToolUsage: r.ToolUsage,
		Tools:     r.Tools,
	}

	for _, opt := range opts {
		opt(req)
	}

	return req
}

type History = []Message

func NewHistory(messages ...Message) History {
	return messages
}

// LLMResponse represents a response from the LLM
type LLMResponse struct {
	Messages []Message
}

// NewLLMResponse creates a new empty LLM response
func NewLLMResponse() *LLMResponse {
	return &LLMResponse{}
}

// AddMessage adds a message to the response
func (r *LLMResponse) AddMessage(message Message) {
	r.Messages = append(r.Messages, message)
}

// AddToolCall adds a tool call to the response
func (r *LLMResponse) AddToolCall(functionCall *ToolCall) {
	r.Messages = append(r.Messages, NewToolCallMessage(functionCall))
}

// CalledTool returns true if the response contains tool calls
func (r *LLMResponse) ToolCalls() []*ToolCall {
	var toolCalls []*ToolCall
	for _, msg := range r.Messages {
		if msg.Kind() == MessageKindToolCall {
			toolCalls = append(toolCalls, msg.(*ToolCallMessage).ToolCall)
		}
	}

	return toolCalls
}
