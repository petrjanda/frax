package llm

import (
	"encoding/json"
)

// Message represents a message in the conversation
type Message interface {
	Kind() MessageKind
	Role() MessageRole
}

// MessageKind represents the type of message
type MessageKind string

const (
	MessageKindText       MessageKind = "text"
	MessageKindToolCall   MessageKind = "tool_call"
	MessageKindToolResult MessageKind = "tool_result"
)

// MessageRole represents the role of the message sender
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
	MessageRoleTool      MessageRole = "tool"
)

// UserMessage represents a message from the user
type UserMessage struct {
	Content string
}

func NewUserMessage(content string) *UserMessage {
	return &UserMessage{
		Content: content,
	}
}

func (m *UserMessage) Kind() MessageKind {
	return MessageKindText
}

func (m *UserMessage) Role() MessageRole {
	return MessageRoleUser
}

// AssistantMessage represents a message from the assistant
type AssistantMessage struct {
	Content string
}

func (m *AssistantMessage) Kind() MessageKind {
	return MessageKindText
}

func (m *AssistantMessage) Role() MessageRole {
	return MessageRoleAssistant
}

// SystemMessage represents a system message
type SystemMessage struct {
	Content string
}

func (m *SystemMessage) Kind() MessageKind {
	return MessageKindText
}

func (m *SystemMessage) Role() MessageRole {
	return MessageRoleSystem
}

type ToolCallMessage struct {
	ToolCall *ToolCall
}

func NewToolCallMessage(toolCall *ToolCall) *ToolCallMessage {
	return &ToolCallMessage{
		ToolCall: toolCall,
	}
}

func (m *ToolCallMessage) Kind() MessageKind {
	return MessageKindToolCall
}

func (m *ToolCallMessage) Role() MessageRole {
	return MessageRoleAssistant
}

// ToolResultMessage represents the result of a tool execution
type ToolResultMessage struct {
	ToolCall *ToolCall
	Result   json.RawMessage
}

func NewToolResultMessage(toolCall *ToolCall, result json.RawMessage) *ToolResultMessage {
	return &ToolResultMessage{
		ToolCall: toolCall,
		Result:   result,
	}
}

func NewToolResultErrorMessage(toolCall *ToolCall, error string) *ToolResultMessage {
	return &ToolResultMessage{
		ToolCall: toolCall,
		Result:   json.RawMessage(error),
	}
}

func (m *ToolResultMessage) Kind() MessageKind {
	return MessageKindToolResult
}

func (m *ToolResultMessage) Role() MessageRole {
	return MessageRoleTool
}

// ToolCall represents a call to a tool
type ToolCall struct {
	ID   string
	Name string
	Args json.RawMessage
}

// ToolErrorMessage represents an error that occurred during tool execution
type ToolErrorMessage struct {
	ToolCall *ToolCall
	Error    string
}

func NewToolErrorMessage(toolCall *ToolCall, error string) *ToolErrorMessage {
	return &ToolErrorMessage{
		ToolCall: toolCall,
		Error:    error,
	}
}

func (m *ToolErrorMessage) Kind() MessageKind {
	return MessageKindToolResult
}

func (m *ToolErrorMessage) Role() MessageRole {
	return MessageRoleAssistant
}
