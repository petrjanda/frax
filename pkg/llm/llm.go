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
	History   []Message
	Tools     []Tool
	ToolUsage bool
}

// LLMRequestOpts represents options for configuring an LLM request
type LLMRequestOpts = func(*LLMRequest)

// WithToolUsage enables or disables tool usage for the request
func WithToolUsage(toolUsage bool) LLMRequestOpts {
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

// NewLLMRequest creates a new LLM request with the given options
func NewLLMRequest(history []Message, opts ...LLMRequestOpts) *LLMRequest {
	r := &LLMRequest{History: history}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

// LLMResponse represents a response from the LLM
type LLMResponse struct {
	Messages  []Message
	ToolCalls []*ToolCall
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
	r.ToolCalls = append(r.ToolCalls, functionCall)
}

// CalledTool returns true if the response contains tool calls
func (r *LLMResponse) CalledTool() bool {
	return len(r.ToolCalls) > 0
}
