package llm

type LLMResponse struct {
	Messages []Message
}

func NewLLMResponse() *LLMResponse {
	return &LLMResponse{}
}

func (r *LLMResponse) AddMessage(message Message) {
	r.Messages = append(r.Messages, message)
}

func (r *LLMResponse) AddToolCall(functionCall *ToolCall) {
	r.Messages = append(r.Messages, NewToolCallMessage(functionCall))
}

func (r *LLMResponse) ToolCalls() []*ToolCall {
	var toolCalls []*ToolCall
	for _, msg := range r.Messages {
		if msg.Kind() == MessageKindToolCall {
			toolCalls = append(toolCalls, msg.(*ToolCallMessage).ToolCall)
		}
	}

	return toolCalls
}
