package llm

import (
	"encoding/json"
)

// Message represents a message in the conversation
type Message interface {
	Kind() MessageKind
}

// MessageKind represents the type of message
type MessageKind string

const (
	MessageKindText       MessageKind = "text"
	MessageKindToolCall   MessageKind = "tool_call"
	MessageKindToolResult MessageKind = "tool_result"
)

// UserMessage represents a message from the user
type UserMessage struct {
	Content string
}

func (m *UserMessage) Kind() MessageKind {
	return MessageKindText
}

// ToolMessage represents a message about a tool
type ToolMessage struct {
	ToolCall *ToolCall
	Result   json.RawMessage
}

func (m *ToolMessage) Kind() MessageKind {
	return MessageKindToolResult
}

// ToolResultMessage represents the result of a tool execution
type ToolResultMessage struct {
	ToolCall *ToolCall
	Result   json.RawMessage
}

func (m *ToolResultMessage) Kind() MessageKind {
	return MessageKindToolResult
}

// ToolCall represents a call to a tool
type ToolCall struct {
	Name string
	Args json.RawMessage
}
