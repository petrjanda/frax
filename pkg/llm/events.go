package llm

import (
	"context"
	"log"
	"log/slog"
	"os"
)

type LLMEvents interface {
	OnRequest(ctx context.Context, request *LLMRequest)
	OnResponse(ctx context.Context, request *LLMRequest, response *LLMResponse)
	OnRequestError(ctx context.Context, request *LLMRequest, err error)
	OnToolError(ctx context.Context, toolCall *ToolCall, attempt int, err error)
}

type AgentEvents interface {
	LLMEvents
}

type NoopAgentEvents struct{}

func (e *NoopAgentEvents) OnRequest(ctx context.Context, request *LLMRequest) {}
func (e *NoopAgentEvents) OnResponse(ctx context.Context, request *LLMRequest, response *LLMResponse) {
}
func (e *NoopAgentEvents) OnRequestError(ctx context.Context, request *LLMRequest, err error) {}
func (e *NoopAgentEvents) OnToolError(ctx context.Context, toolCall *ToolCall, attempt int, err error) {
}

type LogAgentEvents struct {
	logger *slog.Logger
}

func NewLogAgentEvents(logger *slog.Logger) *LogAgentEvents {
	return &LogAgentEvents{logger: logger}
}

func NewJSONFileLogAgentEvents(path string) *LogAgentEvents {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	return &LogAgentEvents{logger: slog.New(slog.NewJSONHandler(file, nil))}
}

func (e *LogAgentEvents) OnRequest(ctx context.Context, request *LLMRequest) {
	e.logger.Info("request",
		"system", request.System,
		"history", request.History,
		"tools", request.Tools,
		"tool_usage", request.ToolUsage,
		"max_completion_tokens", request.MaxCompletionTokens,
		"temperature", request.Temperature,
	)
}

func (e *LogAgentEvents) OnResponse(ctx context.Context, request *LLMRequest, response *LLMResponse) {
	e.logger.Info("response",
		"messages", response.Messages,
	)
}

func (e *LogAgentEvents) OnRequestError(ctx context.Context, request *LLMRequest, err error) {
	e.logger.Error("request error", "error", err)
}

func (e *LogAgentEvents) OnToolError(ctx context.Context, toolCall *ToolCall, attempt int, err error) {
	e.logger.Info("tool call failed",
		"tool", toolCall.Name,
		"attempt", attempt+1,
		"error", err.Error(),
	)
}
