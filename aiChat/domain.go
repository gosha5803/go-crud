package aichat

type AiModelResponse struct {
	Text      string
	IsFinal   bool
	TokensIn  int
	TokensOut int
}

// AiModelRequest — то, что передаём в провайдер.
type AiModelRequest struct {
	// TODO эту модель в теории можно передавать при каждом запросе, типо разная модель
	// Но для моделей разных компаний нужен бужет разный клиент, поэтому пока что отключил этот функционал
	// Model  string // "gemini-3.6-flash", "deepseek-chat" и т.п.
	Prompt string
	// сюда же: SystemPrompt, Temperature, History и т.д.
}

type AiMessageDTO struct {
	Text string `json:"text"`
}

type AiErrorDTO struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type AiDoneDTO struct {
	Reason string `json:"reason,omitempty"`
}
