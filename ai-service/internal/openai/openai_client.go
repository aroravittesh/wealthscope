package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

var (
	chatHTTPClient      *http.Client
	chatCompletionsURL  = "https://api.openai.com/v1/chat/completions"
)

func chatHTTP() *http.Client {
	if chatHTTPClient != nil {
		return chatHTTPClient
	}
	return http.DefaultClient
}

// SetChatHTTPTestConfig swaps the HTTP client and chat-completions URL used by CallOpenAI.
// Pass client nil to keep default client; pass url "" to keep default URL.
func SetChatHTTPTestConfig(client *http.Client, completionsURL string) (cleanup func()) {
	prevC := chatHTTPClient
	prevU := chatCompletionsURL
	if client != nil {
		chatHTTPClient = client
	}
	if strings.TrimSpace(completionsURL) != "" {
		chatCompletionsURL = completionsURL
	}
	return func() {
		chatHTTPClient = prevC
		chatCompletionsURL = prevU
	}
}

func getSystemPrompt() string {
	return `You are the AI assistant for WealthScope, a stock market education and analysis platform.

SCOPE (finance and WealthScope only):
- Help with navigating WealthScope (dashboard, stock search, portfolio, news, risk tools).
- Explain general market concepts, terminology, and high-level risk factors for public equities.
- Refuse unrelated topics (general trivia, politics, sports, coding, entertainment, etc.).

STRICT RULES:
- Do not provide personalized investment advice, buy/sell/hold recommendations, or price targets.
- Do not analyze private portfolio situations beyond what neutral context explicitly states.
- Do not predict exact future prices or guarantee returns.
- If the user only greets you, reply briefly and invite a finance or WealthScope question.
- If unsure the question is on-topic, refuse with exactly:
"I can only assist with stock market and WealthScope-related questions."

GROUNDING:
- The user message may include a "Grounded context" block with labeled sections such as
  [Relevant Financial Knowledge], [Relevant QA Knowledge], [Live Market Data], [News Context], [Portfolio Context], [System Context].
- Base factual claims on those sections when they contain data. Do not invent quotes, prices, or headlines.
- If a section says no data was provided or attached, say clearly that the information is not available in this context (do not guess).

ANSWER FORMAT (when answering a substantive finance question, use this structure):
1. **Explanation** — Short, direct answer to the question using grounded context where present.
2. **Key insight or risk note** — One concise bullet or sentence on uncertainty, limits of the data, or a non-personalized risk angle.
3. **Disclaimer** — End with this exact sentence:
"Investing in the stock market involves risk. WealthScope does not guarantee the accuracy, completeness, or future performance of any information provided and is not responsible for any financial outcomes."

STYLE:
- Be concise, neutral, and factual. Prefer plain language over jargon unless the user uses it first.
- Do not use numbered lists beyond the three-part structure above unless the user asks for a list.`
}

func CallOpenAI(sessionID string, userInput string) (string, error) {
	history := defaultStore.AddUserMessage(sessionID, userInput)

	messages := []Message{
		{Role: "system", Content: getSystemPrompt()},
	}
	messages = append(messages, history...)

	reqBody := OpenAIRequest{
		Model:    "gpt-4o-mini",
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		chatCompletionsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	resp, err := chatHTTP().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("OpenAI API error: " + resp.Status)
	}

	var result OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	reply := result.Choices[0].Message.Content

	defaultStore.AddAssistantMessage(sessionID, reply)

	return reply, nil
}

// ClearSession clears conversation history for a session.
func ClearSession(sessionID string) {
	defaultStore.Clear(sessionID)
}
