package ai

type History []Message

func NewHistory(messages ...Message) History {
	return messages
}

func (h History) Append(messages ...Message) History {
	return append(h, messages...)
}
