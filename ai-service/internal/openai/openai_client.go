package openai

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "os"
    "sync"
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

// ConversationStore holds chat history per session
var (
    conversationStore = make(map[string][]Message)
    mu                sync.Mutex
)

func getSystemPrompt() string {
    return `You are the AI assistant for WealthScope.
WealthScope is a stock market education and analysis platform.
You are ONLY allowed to answer questions related to:
1. Navigating the WealthScope website (dashboard, stock search, portfolio, news, risk analysis).
2. General stock market concepts and terminology.
3. General risk analysis of publicly traded stocks.
STRICT RULES:
- If a question is unrelated to stock markets or WealthScope, politely refuse.
- Do NOT answer general knowledge, entertainment, politics, sports, coding, or random questions.
- Do NOT analyze user portfolios or personal financial situations.
- Do NOT provide personalized investment advice, buy/sell recommendations, or price targets.
- Do NOT predict exact future stock prices.
- If unsure whether the question is stock-related, refuse.
- If the user sends a greeting (e.g., "hi", "hello"), respond briefly and professionally,
  then guide them to ask a stock market or WealthScope-related question.
- Do NOT engage in casual conversation or small talk.
When refusing, respond exactly with:
"I can only assist with stock market and WealthScope-related questions."
For stock risk analysis:
- Discuss volatility
- Industry and sector risks
- Market and macroeconomic conditions
- Competitive landscape
- General company fundamentals (high-level only)
Keep analysis neutral and informational.
Always include this disclaimer at the end:
"Investing in the stock market involves risk. WealthScope does not guarantee the accuracy,
completeness, or future performance of any information provided and is not responsible
for any financial outcomes."
Keep responses clear, structured, and professional.`
}

func CallOpenAI(sessionID string, userInput string) (string, error) {
    mu.Lock()

    // Initialize session if new
    if _, exists := conversationStore[sessionID]; !exists {
        conversationStore[sessionID] = []Message{}
    }

    // Append user message to history
    conversationStore[sessionID] = append(conversationStore[sessionID], Message{
        Role:    "user",
        Content: userInput,
    })

    // Build messages: system prompt + full history
    messages := []Message{
        {Role: "system", Content: getSystemPrompt()},
    }
    messages = append(messages, conversationStore[sessionID]...)
    mu.Unlock()

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
        "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
    if err != nil {
        return "", err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

    client := &http.Client{}
    resp, err := client.Do(req)
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

    // Save assistant reply to history
    mu.Lock()
    conversationStore[sessionID] = append(conversationStore[sessionID], Message{
        Role:    "assistant",
        Content: reply,
    })
    // Keep last 10 messages to avoid token overflow
    if len(conversationStore[sessionID]) > 10 {
        conversationStore[sessionID] = conversationStore[sessionID][len(conversationStore[sessionID])-10:]
    }
    mu.Unlock()

    return reply, nil
}

// ClearSession clears conversation history for a session
func ClearSession(sessionID string) {
    mu.Lock()
    defer mu.Unlock()
    delete(conversationStore, sessionID)
}