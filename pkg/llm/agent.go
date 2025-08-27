package llm

import (
	"context"
	"fmt"
)

// IAgent represents an agent that can process conversations
type IAgent interface {
	Loop(ctx context.Context, history []Message) ([]Message, error)
}

// Agent represents an agent that can use tools and interact with an LLM
type Agent struct {
	llm     LLM
	tools   []Tool
	history []Message
}

// AgentOpts represents options for configuring an agent
type AgentOpts = func(*Agent)

// WithInitialHistory sets the initial conversation history for the agent
func WithInitialHistory(history []Message) AgentOpts {
	return func(a *Agent) {
		a.history = history
	}
}

// NewAgent creates a new agent with the given LLM and tools
func NewAgent(llm LLM, tools []Tool, opts ...AgentOpts) IAgent {
	a := &Agent{llm: llm, tools: tools}
	for _, opt := range opts {
		opt(a)
	}

	return a
}

// Loop processes the conversation loop, handling tool calls and LLM responses
func (a *Agent) Loop(ctx context.Context, history []Message) ([]Message, error) {
	req := NewLLMRequest(history,
		WithTools(a.tools...),
		WithToolUsage(AutoToolSelection()),
	)

	response, err := a.llm.Invoke(ctx, req)
	if err != nil {
		return nil, err
	}

	if response.CalledTool() {
		for _, toolCall := range response.ToolCalls {
			message, err := a.CallTool(ctx, toolCall)
			if err != nil {
				return nil, err
			}

			response.AddMessage(message)
		}

		// Continue the loop
		return a.Loop(ctx, response.Messages)
	}

	return response.Messages, nil
}

// CallTool executes a tool call and returns the result message
func (a *Agent) CallTool(ctx context.Context, toolCall *ToolCall) (Message, error) {
	var tool Tool
	for _, t := range a.tools {
		if t.Name() == toolCall.Name {
			tool = t
			break
		}
	}

	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolCall.Name)
	}

	result, err := tool.Run(ctx, toolCall.Args)
	if err != nil {
		return nil, err
	}

	return &ToolResultMessage{
		ToolCall: toolCall,
		Result:   result,
	}, nil
}
