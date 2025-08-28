package openai

import (
	"fmt"

	"github.com/openai/openai-go/v2"
	"github.com/petrjanda/frax/pkg/llm"
)

// convertToolUsage converts our ToolUsage interface to OpenAI's tool choice format
// Returns nil when no specific tool choice is needed (auto/default behavior)
func convertToolUsage(toolUsage llm.ToolUsage, tools []llm.Tool) (*openai.ChatCompletionToolChoiceOptionUnionParam, error) {
	switch toolUsage.Type() {
	default:
		return nil, nil

	case llm.ToolUsageAuto:
		return nil, nil

	case llm.ToolUsageForced:
		if forced, ok := toolUsage.(*llm.ForcedToolUsage); ok {
			tool, err := findTool(forced.ToolName, tools)
			if err != nil {
				return nil, fmt.Errorf("forced tool %s not available", forced.ToolName)
			}

			toolChoice := openai.ToolChoiceOptionFunctionToolChoice(openai.ChatCompletionNamedToolChoiceFunctionParam{
				Name: tool.Name(),
			})
			return &toolChoice, nil
		}

		return nil, nil
	}
}

func findTool(name string, tools []llm.Tool) (llm.Tool, error) {
	for _, tool := range tools {
		if tool.Name() == name {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool %s not found", name)
}
