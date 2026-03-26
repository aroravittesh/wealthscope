package handler

import (
    "net/http"

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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }
    if body.Message == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
        return
    }

    sessionID := body.SessionID
    if sessionID == "" {
        sessionID = "default"
    }

    response, err := service.ProcessMessage(sessionID, body.Message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
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
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }
        if body.Message == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
            return
        }

        sessionID := body.SessionID
        if sessionID == "" {
            sessionID = "default"
        }

        response, err := svc.ProcessMessage(sessionID, body.Message)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "response":   response,
            "session_id": sessionID,
        })
    }
}

func ClearChatHandler(c *gin.Context) {
    sessionID := c.Param("session_id")
    if sessionID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
        return
    }
    openai.ClearSession(sessionID)
    c.JSON(http.StatusOK, gin.H{"message": "Session cleared"})
}