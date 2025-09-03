package llm

type LLMResponse struct {
	Messages History   `json:"messages"`
	Usage    *LLMUsage `json:"usage"`
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

func (r *LLMResponse) Clone() *LLMResponse {
	messages := make(History, len(r.Messages))
	copy(messages, r.Messages)

	return &LLMResponse{
		Messages: messages,
		Usage:    r.Usage,
	}
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

func (r *LLMResponse) SetUsage(usage *LLMUsage) {
	r.Usage = usage
}

type LLMUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

func NewLLMUsage(promptTokens, completionTokens, totalTokens int64) *LLMUsage {
	return &LLMUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}
}

func (u *LLMUsage) Add(other *LLMUsage) {
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.TotalTokens += other.TotalTokens
}
