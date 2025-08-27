package openai

import (
	"testing"

	"github.com/petrjanda/frax/pkg/llm"
)

func TestConvertToolUsage(t *testing.T) {
	tests := []struct {
		name      string
		toolUsage llm.ToolUsage
		tools     []llm.Tool
		expectNil bool
	}{
		{
			name:      "Auto tool selection should return nil",
			toolUsage: llm.AutoToolSelection(),
			tools:     []llm.Tool{},
			expectNil: true,
		},
		{
			name:      "Forced tool should return tool choice",
			toolUsage: llm.ForceTool("calculator"),
			tools:     []llm.Tool{},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToolUsage(tt.toolUsage, tt.tools)

			if tt.expectNil && result != nil {
				t.Errorf("expected nil result, got %v", result)
			}

			if !tt.expectNil && result == nil {
				t.Errorf("expected non-nil result, got nil")
			}
		})
	}
}
