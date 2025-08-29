package llm

type LLMResponse struct {
	Messages History
}

func NewLLMResponse() *LLMResponse {
	return &LLMResponse{}
}

func (r *LLMResponse) AddMessage(message Message) {
	r.Messages = r.Messages.Append(message)
}

func (r *LLMResponse) AddToolCall(functionCall *ToolCall) {
	r.Messages = r.Messages.Append(NewToolCallMessage(functionCall))
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
