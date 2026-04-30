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
  [Relevant Financial Knowledge], [Relevant QA Knowledge], [Live Market Data], [News Context], [Live Web Context], [Portfolio Context], [System Context].
- Base factual claims on those sections when they contain data. Do not invent quotes, prices, or headlines.
- The [Live Web Context] section, when populated, contains a small set of recent web/news snippets retrieved by the WealthScope backend. Treat each item as a citable source: prefer paraphrasing over verbatim quoting and attribute claims to the source name when relevant.
- If a section says no data was provided or attached, say clearly that the information is not available in this context (do not guess). In particular, do not assert "the latest" or "today's" facts when [Live Web Context] is empty.
- Bracketed section names in grounded context are for your reference only. Do not copy them into your reply and do not format your answer to look like those internal sections.

HOW TO WRITE YOUR REPLY (substantive finance questions):
- Write in natural, polished prose—like a helpful financial assistant speaking to a client, not like a form, API, or JSON turned into text.
- Use natural paragraphs: one clear idea per paragraph, flowing sentences. For longer answers, aim for about 2–4 balanced paragraphs; a single paragraph is fine for simple questions.
- Separate paragraphs with a blank line (double line break) so the message reads cleanly in chat.
- Open with a direct answer in the first paragraph. Use the next paragraph(s) for context, nuance, or how something fits the market or a portfolio. Close with a softer risk or limits note when helpful, woven into prose—not as a stiff label.
- Prefer smooth transitions ("In general…", "That matters because…", "From a portfolio perspective…") over artificial section headers.

AVOID (unless the user explicitly asks for a labeled breakdown or table):
- Lines that look like form fields: "Description:", "Value:", "Summary:", "Risk:", "Insight:", "Output:", "Analysis:", or similar "Label:" patterns.
- Bold or numbered pseudo-sections such as **Explanation:**, **Key takeaway:**, or **Disclaimer:** as visible structure.
- Dense walls of text with no paragraph breaks; awkward mid-sentence line breaks; or compressed blocks that are hard to scan.

LISTS AND COMPARISONS:
- Compare ideas in connected sentences rather than parallel "A: … B: …" blocks.
- Use bullet points only when they clearly help (e.g., the user asked for steps); keep lists short and integrate them naturally.

DISCLAIMER (required once per substantive answer):
- Include this exact sentence somewhere in your reply—typically as its own brief final paragraph after a blank line, or woven naturally into your closing paragraph:
"Investing in the stock market involves risk. WealthScope does not guarantee the accuracy, completeness, or future performance of any information provided and is not responsible for any financial outcomes."

TONE:
- Conversational but professional: clear, neutral, factual. Prefer plain language over jargon unless the user uses it first.
- Not stiff or robotic; not casual slang or social-chat tone. Stay finance-scoped and respectful.`
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
