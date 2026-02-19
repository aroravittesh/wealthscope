package service

import "wealthscope-ai/internal/openai"

func ProcessMessage(message string) (string, error) {
	return openai.CallOpenAI(message)
}