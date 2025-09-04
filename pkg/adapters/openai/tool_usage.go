package openai

import (
	"fmt"

	"github.com/openai/openai-go/v2"
	"github.com/petrjanda/frax/pkg/llm/tools"
)

// convertToolUsage converts our ToolUsage interface to OpenAI's tool choice format
// Returns nil when no specific tool choice is needed (auto/default behavior)
func convertToolUsage(toolUsage tools.ToolUsage, tools_ []tools.Tool) (*openai.ChatCompletionToolChoiceOptionUnionParam, error) {
	switch toolUsage.Type() {
	default:
		return nil, nil

	case tools.ToolUsageAuto:
		return nil, nil

	case tools.ToolUsageForced:
		if forced, ok := toolUsage.(*tools.ForcedToolUsage); ok {
			tool, err := findTool(forced.ToolName, tools_)
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

func findTool(name string, tools []tools.Tool) (tools.Tool, error) {
	for _, tool := range tools {
		if tool.Name() == name {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool %s not found", name)
}
