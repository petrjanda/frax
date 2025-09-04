package llm

import "fmt"

type Toolbox = []Tool

func FindTool(name string, tools Toolbox) (Tool, error) {
	for _, tool := range tools {
		if tool.Name() == name {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", name)
}

func NewToolbox(tools ...Tool) Toolbox {
	return tools
}
