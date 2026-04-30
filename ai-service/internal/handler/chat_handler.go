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

	if err := c.BindJSON(&body); err != nil {
			RespondBadRequest(c, "Request failed", "Invalid request")
		return
	}
	if body.Message == "" {
			RespondBadRequest(c, "Request failed", "Message cannot be empty")
		return
	}

	sessionID := body.SessionID
	if sessionID == "" {
		sessionID = "default"
	}

	service.LogChatRequestStart(sessionID, body.Message)

	response, err := service.ProcessMessage(sessionID, body.Message)
	if err != nil {
		service.LogChatRequestFailed(sessionID, body.Message, err)
		RespondError(c, 500, "Request failed", err.Error())
		return
	}

	service.LogChatRequestComplete(sessionID, body.Message, response)
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

        if err := c.BindJSON(&body); err != nil {
			RespondBadRequest(c, "Request failed", "Invalid request")
            return
        }
        if body.Message == "" {
			RespondBadRequest(c, "Request failed", "Message cannot be empty")
            return
        }

        sessionID := body.SessionID
        if sessionID == "" {
            sessionID = "default"
        }

        response, err := svc.ProcessMessage(sessionID, body.Message)
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
    sessionID := c.Param("session_id")
    if sessionID == "" {
		RespondBadRequest(c, "Request failed", "session_id required")
        return
    }
    openai.ClearSession(sessionID)
	RespondSuccess(c, 200, "Session cleared", gin.H{"session_id": sessionID})
}