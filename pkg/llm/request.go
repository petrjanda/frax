package llm

type LLMRequest struct {
	System    string
	History   History
	Tools     []Tool
	ToolUsage ToolUsage

	MaxCompletionTokens int
	Temperature         float64
}

type LLMRequestOpts = func(*LLMRequest)

func WithToolUsage(toolUsage ToolUsage) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.ToolUsage = toolUsage
	}
}

func WithTools(tools ...Tool) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.Tools = append(r.Tools, tools...)
	}
}

func WithSystem(system string) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.System = system
	}
}

func WithHistory(history History) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.History = history
	}
}

func WithMaxCompletionTokens(maxCompletionTokens int) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.MaxCompletionTokens = maxCompletionTokens
	}
}

func WithTemperature(temperature float64) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.Temperature = temperature
	}
}

func NewLLMRequest(history History, opts ...LLMRequestOpts) *LLMRequest {
	r := &LLMRequest{
		History:   history,
		ToolUsage: AutoToolSelection(), // Default to auto tool selection
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *LLMRequest) Clone(opts ...LLMRequestOpts) *LLMRequest {
	req := &LLMRequest{
		History:             r.History,
		ToolUsage:           r.ToolUsage,
		Tools:               r.Tools,
		System:              r.System,
		MaxCompletionTokens: r.MaxCompletionTokens,
		Temperature:         r.Temperature,
	}

	for _, opt := range opts {
		opt(req)
	}

	return req
}
