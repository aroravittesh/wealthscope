package handler

import (
	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/openai"
	"wealthscope-ai/internal/service"
)

func ChatHandler(c *gin.Context) {
	var body struct {
		Message   string `json:"message"`
		SessionID string `json:"session_id"`
	}

	if !BindJSONOrRespond(c, &body, "Invalid request") {
		return
	}
	message, ok := RequiredTrimmed(c, body.Message, "Message cannot be empty")
	if !ok {
		return
	}

	sessionID := NormalizeSessionID(body.SessionID)

	service.LogChatRequestStart(sessionID, message)

	response, err := service.ProcessMessage(sessionID, message)
	if err != nil {
		service.LogChatRequestFailed(sessionID, message, err)
		RespondError(c, 500, "Request failed", err.Error())
		return
	}

	service.LogChatRequestComplete(sessionID, message, response)
	RespondSuccess(c, 200, "Chat response generated", gin.H{
		"response":   response,
		"session_id": sessionID,
	})
}

// ChatHandlerWithService allows injecting a mock service for unit testing
func ChatHandlerWithService(svc service.ChatServiceInterface) gin.HandlerFunc {
    return func(c *gin.Context) {
        var body struct {
            Message   string `json:"message"`
            SessionID string `json:"session_id"`
        }

		if !BindJSONOrRespond(c, &body, "Invalid request") {
            return
        }
		message, ok := RequiredTrimmed(c, body.Message, "Message cannot be empty")
		if !ok {
            return
        }

		sessionID := NormalizeSessionID(body.SessionID)

		response, err := svc.ProcessMessage(sessionID, message)
        if err != nil {
			RespondError(c, 500, "Request failed", err.Error())
            return
        }

		RespondSuccess(c, 200, "Chat response generated", gin.H{
            "response":   response,
            "session_id": sessionID,
        })
    }
}

func ClearChatHandler(c *gin.Context) {
	sessionID, ok := RequirePathParam(c, "session_id", "session_id required")
	if !ok {
		return
	}
	openai.ClearSession(sessionID)
	RespondSuccess(c, 200, "Session cleared", gin.H{"session_id": sessionID})
}