package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"neobase-ai/internal/apis/dtos"
	"neobase-ai/internal/services"
	"neobase-ai/internal/utils"
	"net/http"
	"strconv"
	"sync"
	"time"

	"neobase-ai/pkg/dbmanager"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService services.ChatService
	dbManager   *dbmanager.Manager
	streamMutex sync.RWMutex
	streams     map[string]chan dtos.StreamResponse // key: userID:chatID:streamID
}

func NewChatHandler(chatService services.ChatService, dbManager *dbmanager.Manager) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		dbManager:   dbManager,
		streamMutex: sync.RWMutex{},
		streams:     make(map[string]chan dtos.StreamResponse),
	}
}

func (h *ChatHandler) Create(c *gin.Context) {
	var req dtos.CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorMsg := err.Error()
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	userID := c.GetString("userID")
	response, statusCode, err := h.chatService.Create(userID, &req)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    response,
	})
}

func (h *ChatHandler) List(c *gin.Context) {
	userID := c.GetString("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, statusCode, err := h.chatService.List(userID, page, pageSize)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    response,
	})
}

func (h *ChatHandler) GetByID(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")

	response, statusCode, err := h.chatService.GetByID(userID, chatID)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    response,
	})
}

func (h *ChatHandler) Update(c *gin.Context) {
	var req dtos.UpdateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorMsg := err.Error()
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	userID := c.GetString("userID")
	chatID := c.Param("id")

	response, statusCode, err := h.chatService.Update(userID, chatID, &req)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    response,
	})
}

func (h *ChatHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")

	statusCode, err := h.chatService.Delete(userID, chatID)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    "Chat deleted successfully",
	})
}

func (h *ChatHandler) ListMessages(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	response, statusCode, err := h.chatService.ListMessages(userID, chatID, page, pageSize)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    response,
	})
}

func (h *ChatHandler) CreateMessage(c *gin.Context) {
	var req dtos.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorMsg := err.Error()
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	userID := c.GetString("userID")
	chatID := c.Param("id")

	response, statusCode, err := h.chatService.CreateMessage(c.Request.Context(), userID, chatID, req.StreamID, req.Content)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    response,
	})
}

func (h *ChatHandler) UpdateMessage(c *gin.Context) {
	var req dtos.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorMsg := err.Error()
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	userID := c.GetString("userID")
	chatID := c.Param("id")
	messageID := c.Param("messageId")

	response, statusCode, err := h.chatService.UpdateMessage(userID, chatID, messageID, req.StreamID, &req)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    response,
	})
}

func (h *ChatHandler) DeleteMessages(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")

	statusCode, err := h.chatService.DeleteMessages(userID, chatID)
	if err != nil {
		errorMsg := err.Error()
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	c.JSON(int(statusCode), dtos.Response{
		Success: true,
		Data:    "Messages deleted successfully",
	})
}

// HandleStreamEvent implements the StreamHandler interface
func (h *ChatHandler) HandleStreamEvent(userID, chatID, streamID string, response dtos.StreamResponse) {
	streamKey := fmt.Sprintf("%s:%s:%s", userID, chatID, streamID)

	h.streamMutex.RLock()
	streamChan, exists := h.streams[streamKey]
	h.streamMutex.RUnlock()

	if !exists {
		log.Printf("No stream found for key: %s", streamKey)
		return
	}

	// Try to send with timeout
	select {
	case streamChan <- response:
		log.Printf("Successfully sent event to stream: %s, event: %s", streamKey, response.Event)
	case <-time.After(100 * time.Millisecond):
		log.Printf("Timeout sending event to stream: %s", streamKey)
	}
}

// StreamChat handles SSE endpoint
func (h *ChatHandler) StreamChat(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")
	streamID := c.Query("stream_id")

	if streamID == "" {
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr("stream_id is required"),
		})
		return
	}

	streamKey := fmt.Sprintf("%s:%s:%s", userID, chatID, streamID)
	log.Printf("Starting stream for key: %s", streamKey)

	// Create buffered channel
	h.streamMutex.Lock()
	streamChan := make(chan dtos.StreamResponse, 100)
	h.streams[streamKey] = streamChan
	h.streamMutex.Unlock()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	ctx := c.Request.Context()
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Set the stream handler
	h.chatService.SetStreamHandler(h)

	// Cleanup on exit
	defer func() {
		h.streamMutex.Lock()
		if ch, exists := h.streams[streamKey]; exists {
			close(ch)
			delete(h.streams, streamKey)
			log.Printf("Cleaned up stream for key: %s", streamKey)
		}
		h.streamMutex.Unlock()
	}()

	// Send initial connection event
	data, _ := json.Marshal(dtos.StreamResponse{
		Event: "connected",
		Data:  "Stream established",
	})
	c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
	c.Writer.Flush()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Client disconnected for stream key: %s", streamKey)
			return

		case <-heartbeatTicker.C:
			data, _ := json.Marshal(dtos.StreamResponse{
				Event: "heartbeat",
				Data:  "ping",
			})
			c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
			c.Writer.Flush()

		case msg, ok := <-streamChan:
			if !ok {
				log.Printf("Stream channel closed for key: %s", streamKey)
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}
			log.Printf("Sending stream event -> key: %s, event: %s", streamKey, msg.Event)
			c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
			c.Writer.Flush()
		}
	}
}

// CancelStream cancels currently streaming response
func (h *ChatHandler) CancelStream(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")
	streamID := c.Query("stream_id")

	if streamID == "" {
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr("stream_id is required"),
		})
		return
	}

	// Create stream key
	streamKey := fmt.Sprintf("%s:%s:%s", userID, chatID, streamID)

	// First cancel the processing
	h.chatService.CancelProcessing(userID, chatID, streamID)

	// Then cleanup the stream
	h.streamMutex.Lock()
	if streamChan, ok := h.streams[streamKey]; ok {
		close(streamChan)
		delete(h.streams, streamKey)
	}
	h.streamMutex.Unlock()

	c.JSON(http.StatusOK, dtos.Response{
		Success: true,
		Data:    "Operation cancelled successfully",
	})
}

// ConnectDB establishes a database connection
func (h *ChatHandler) ConnectDB(c *gin.Context) {

	var req dtos.ConnectDBRequest
	userID := c.GetString("userID")
	chatID := c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(fmt.Sprintf("Invalid request: %v", err)),
		})
		return
	}

	statusCode, err := h.chatService.ConnectDB(c.Request.Context(), userID, chatID, req.StreamID)
	if err != nil {
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, dtos.Response{
		Success: true,
		Data:    "Database connected successfully",
	})
}

// DisconnectDB closes a database connection
func (h *ChatHandler) DisconnectDB(c *gin.Context) {
	var req dtos.DisconnectDBRequest
	userID := c.GetString("userID")
	chatID := c.Param("id")
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(fmt.Sprintf("Invalid request: %v", err)),
		})
		return
	}

	statusCode, err := h.chatService.DisconnectDB(c.Request.Context(), userID, chatID, req.StreamID)
	if err != nil {
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, dtos.Response{
		Success: true,
		Data:    "Database disconnected successfully",
	})
}

// GetDBConnectionStatus checks the current connection status
func (h *ChatHandler) GetDBConnectionStatus(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")

	status, statusCode, err := h.chatService.GetDBConnectionStatus(c.Request.Context(), userID, chatID)
	if err != nil {
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, dtos.Response{
		Success: true,
		Data:    status,
	})
}

func (h *ChatHandler) RefreshSchema(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")

	statusCode, err := h.chatService.RefreshSchema(c.Request.Context(), userID, chatID)
	if err != nil {
		c.JSON(int(statusCode), dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(err.Error()),
		})
	}

	c.JSON(http.StatusOK, dtos.Response{
		Success: true,
		Data:    "Schema refreshed successfully",
	})
}

// Add query execution methods
func (h *ChatHandler) ExecuteQuery(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")

	var req dtos.ExecuteQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(err.Error()),
		})
		return
	}

	// Execute query
	response, status, err := h.chatService.ExecuteQuery(c.Request.Context(), userID, chatID, &req)
	if err != nil {
		c.JSON(int(status), dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(err.Error()),
		})
		return
	}

	c.JSON(int(status), response)
}

func (h *ChatHandler) RollbackQuery(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")

	var req dtos.RollbackQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(err.Error()),
		})
		return
	}

	// Execute rollback
	response, status, err := h.chatService.RollbackQuery(c.Request.Context(), userID, chatID, &req)
	if err != nil {
		c.JSON(int(status), dtos.Response{
			Success: false,
			Error:   utils.ToStringPtr(err.Error()),
		})
		return
	}

	c.JSON(int(status), response)
}

func (h *ChatHandler) CancelQueryExecution(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")
	var req dtos.CancelQueryExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cancel execution
	h.chatService.CancelQueryExecution(userID, chatID, req.MessageID, req.QueryID, req.StreamID)
	c.JSON(http.StatusOK, dtos.Response{
		Success: true,
		Data:    "Query execution cancelled successfully",
	})
}

// Update the stream handling
func (h *ChatHandler) HandleStream(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("id")
	streamID := c.Query("stream_id")

	streamKey := fmt.Sprintf("%s:%s:%s", userID, chatID, streamID)
	log.Printf("Starting stream for key: %s", streamKey)

	// Create buffered channel
	h.streamMutex.Lock()
	streamChan := make(chan dtos.StreamResponse, 100)
	h.streams[streamKey] = streamChan
	h.streamMutex.Unlock()

	log.Printf("Created new stream channel for key: %s", streamKey)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	ctx := c.Request.Context()
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Cleanup on exit
	defer func() {
		h.streamMutex.Lock()
		if ch, exists := h.streams[streamKey]; exists {
			close(ch)
			delete(h.streams, streamKey)
			log.Printf("Cleaned up stream for key: %s", streamKey)
		}
		h.streamMutex.Unlock()
	}()

	// Send initial connection event
	c.SSEvent("message", dtos.StreamResponse{
		Event: "connected",
		Data:  "Stream established",
	})
	c.Writer.Flush()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Client disconnected for stream key: %s", streamKey)
			return

		case <-heartbeatTicker.C:
			c.SSEvent("message", dtos.StreamResponse{
				Event: "heartbeat",
				Data:  "ping",
			})
			c.Writer.Flush()

		case msg, ok := <-streamChan:
			if !ok {
				log.Printf("Stream channel closed for key: %s", streamKey)
				return
			}
			log.Printf("Sending stream event -> key: %s, event: %s", streamKey, msg.Event)
			c.SSEvent("message", msg)
			c.Writer.Flush()
		}
	}
}
