package openai

import (
	"github.com/openai/openai-go/v2"
	"github.com/petrjanda/frax/pkg/llm"
)

// convertToolUsage converts our ToolUsage interface to OpenAI's tool choice format
// Returns nil when no specific tool choice is needed (auto/default behavior)
func convertToolUsage(toolUsage llm.ToolUsage, tools []llm.Tool) *openai.ChatCompletionToolChoiceOptionUnionParam {
	switch toolUsage.Type() {
	case llm.ToolUsageAuto:
		// For auto tool selection, we don't set ToolChoice
		// This lets OpenAI use the default behavior
		return nil

	case llm.ToolUsageForced:
		// Force use of a specific tool
		if forced, ok := toolUsage.(*llm.ForcedToolUsage); ok {
			toolChoice := openai.ToolChoiceOptionFunctionToolChoice(openai.ChatCompletionNamedToolChoiceFunctionParam{
				Name: forced.ToolName,
			})
			return &toolChoice
		}
		// Fallback to no specific choice (auto)
		return nil

	default:
		// Default to no specific choice (auto)
		return nil
	}
}
