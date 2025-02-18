package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"neobase-ai/internal/apis/dtos"
	"neobase-ai/internal/constants"
	"neobase-ai/internal/models"
	"neobase-ai/internal/repositories"
	"neobase-ai/internal/utils"
	"neobase-ai/pkg/dbmanager"
	"neobase-ai/pkg/llm"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Used by Handler
type StreamHandler interface {
	HandleStreamEvent(userID, chatID, streamID string, response dtos.StreamResponse)
}

type ChatService interface {
	Create(userID string, req *dtos.CreateChatRequest) (*dtos.ChatResponse, uint32, error)
	Update(userID, chatID string, req *dtos.UpdateChatRequest) (*dtos.ChatResponse, uint32, error)
	Delete(userID, chatID string) (uint32, error)
	GetByID(userID, chatID string) (*dtos.ChatResponse, uint32, error)
	List(userID string, page, pageSize int) (*dtos.ChatListResponse, uint32, error)
	CreateMessage(ctx context.Context, userID, chatID string, streamID string, content string) (*dtos.MessageResponse, uint16, error)
	UpdateMessage(userID, chatID, messageID string, req *dtos.CreateMessageRequest) (*dtos.MessageResponse, uint32, error)
	DeleteMessages(userID, chatID string) (uint32, error)
	ListMessages(userID, chatID string, page, pageSize int) (*dtos.MessageListResponse, uint32, error)
	SetStreamHandler(handler StreamHandler)
	CancelProcessing(streamID string)
	ConnectDB(ctx context.Context, userID, chatID string, streamID string) (uint32, error)
	DisconnectDB(ctx context.Context, userID, chatID string, streamID string) (uint32, error)
	GetDBConnectionStatus(ctx context.Context, userID, chatID string) (*dtos.ConnectionStatusResponse, uint32, error)
	HandleDBEvent(userID, chatID, streamID string, response dtos.StreamResponse)
}

type chatService struct {
	chatRepo        repositories.ChatRepository
	llmRepo         repositories.LLMMessageRepository
	dbManager       *dbmanager.Manager
	llmClient       llm.Client
	streamChans     map[string]chan dtos.StreamResponse
	streamHandler   StreamHandler
	activeProcesses map[string]context.CancelFunc // key: streamID
	processesMu     sync.RWMutex
}

func NewChatService(
	chatRepo repositories.ChatRepository,
	llmRepo repositories.LLMMessageRepository,
	dbManager *dbmanager.Manager,
	llmClient llm.Client,
) ChatService {
	return &chatService{
		chatRepo:        chatRepo,
		llmRepo:         llmRepo,
		dbManager:       dbManager,
		llmClient:       llmClient,
		streamChans:     make(map[string]chan dtos.StreamResponse),
		activeProcesses: make(map[string]context.CancelFunc),
	}
}

func (s *chatService) Create(userID string, req *dtos.CreateChatRequest) (*dtos.ChatResponse, uint32, error) {
	log.Printf("Creating chat for user %s", userID)
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid user ID format")
	}

	// Create connection object
	connection := models.Connection{
		Type:     req.Connection.Type,
		Host:     req.Connection.Host,
		Port:     req.Connection.Port,
		Username: &req.Connection.Username,
		Password: &req.Connection.Password,
		Database: req.Connection.Database,
		Base:     models.NewBase(),
	}

	// Create chat with connection
	chat := models.NewChat(userObjID, connection)
	if err := s.chatRepo.Create(chat); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create chat: %v", err)
	}

	return s.buildChatResponse(chat), http.StatusCreated, nil
}

func (s *chatService) Update(userID, chatID string, req *dtos.UpdateChatRequest) (*dtos.ChatResponse, uint32, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid user ID format")
	}

	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid chat ID format")
	}

	// Get existing chat
	chat, err := s.chatRepo.FindByID(chatObjID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch chat: %v", err)
	}
	if chat == nil {
		return nil, http.StatusNotFound, fmt.Errorf("chat not found")
	}
	if chat.UserID != userObjID {
		return nil, http.StatusForbidden, fmt.Errorf("unauthorized access to chat")
	}

	// Update chat fields
	if req.Connection != (dtos.CreateConnectionRequest{}) {
		chat.Connection = models.Connection{
			Type:     req.Connection.Type,
			Host:     req.Connection.Host,
			Port:     req.Connection.Port,
			Username: &req.Connection.Username,
			Password: &req.Connection.Password,
			Database: req.Connection.Database,
			Base:     models.NewBase(),
		}
	}

	if err := s.chatRepo.Update(chatObjID, chat); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to update chat: %v", err)
	}

	return s.buildChatResponse(chat), http.StatusOK, nil
}

func (s *chatService) Delete(userID, chatID string) (uint32, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid user ID format")
	}

	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid chat ID format")
	}

	// Verify ownership
	chat, err := s.chatRepo.FindByID(chatObjID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to fetch chat: %v", err)
	}
	if chat == nil {
		return http.StatusNotFound, fmt.Errorf("chat not found")
	}
	if chat.UserID != userObjID {
		return http.StatusForbidden, fmt.Errorf("unauthorized access to chat")
	}

	// Delete chat and its messages
	if err := s.chatRepo.Delete(chatObjID); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete chat: %v", err)
	}
	if err := s.llmRepo.DeleteMessagesByChatID(chatObjID); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete chat messages: %v", err)
	}

	return http.StatusOK, nil
}

func (s *chatService) GetByID(userID, chatID string) (*dtos.ChatResponse, uint32, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid user ID format")
	}

	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid chat ID format")
	}

	chat, err := s.chatRepo.FindByID(chatObjID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch chat: %v", err)
	}
	if chat == nil {
		return nil, http.StatusNotFound, fmt.Errorf("chat not found")
	}
	if chat.UserID != userObjID {
		return nil, http.StatusForbidden, fmt.Errorf("unauthorized access to chat")
	}

	return s.buildChatResponse(chat), http.StatusOK, nil
}

func (s *chatService) List(userID string, page, pageSize int) (*dtos.ChatListResponse, uint32, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid user ID format")
	}

	chats, total, err := s.chatRepo.FindByUserID(userObjID, page, pageSize)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch chats: %v", err)
	}

	response := &dtos.ChatListResponse{
		Chats: make([]dtos.ChatResponse, len(chats)),
		Total: total,
	}

	for i, chat := range chats {
		response.Chats[i] = *s.buildChatResponse(chat)
	}

	return response, http.StatusOK, nil
}

// Add new types for message handling
type MessageType string

const (
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeSystem    MessageType = "system"
)

// Update CreateMessage to use a separate context for LLM processing
func (s *chatService) CreateMessage(ctx context.Context, userID, chatID string, streamID string, content string) (*dtos.MessageResponse, uint16, error) {
	log.Printf("CreateMessage -> userID: %s, chatID: %s, streamID: %s, content: %s", userID, chatID, streamID, content)

	// 1. Save user message
	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid chat ID format")
	}

	msg := &models.Message{
		Base:    models.NewBase(),
		ChatID:  chatObjID,
		Content: content,
		Type:    string(MessageTypeUser),
	}

	log.Printf("CreateMessage -> msg: %+v", msg)

	if err := s.chatRepo.CreateMessage(msg); err != nil {
		log.Printf("CreateMessage -> Error saving message: %v", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save message: %v", err)
	}

	log.Printf("CreateMessage -> msg saved")

	// 2. Save to LLM messages for context
	llmMsg := &models.LLMMessage{
		Base:   models.NewBase(),
		ChatID: chatObjID,
		Role:   "user",
		Content: map[string]interface{}{
			"user_message": content,
		},
	}

	if err := s.llmRepo.CreateMessage(llmMsg); err != nil {
		log.Printf("CreateMessage -> Error saving LLM message: %v", err)
		// Continue even if LLM message save fails
	}

	log.Printf("CreateMessage -> llmMsg saved")

	// 3. Start async LLM processing with a background context
	go func() {
		// Create a new background context for LLM processing
		bgCtx := context.Background()
		s.processLLMResponse(bgCtx, userID, msg.ID.Hex(), chatID, streamID)
	}()

	return s.buildMessageResponse(msg), http.StatusOK, nil
}

// Update processLLMResponse to handle the new context
func (s *chatService) processLLMResponse(ctx context.Context, userID, messageID, chatID, streamID string) {
	log.Printf("processLLMResponse -> userID: %s, chatID: %s, streamID: %s", userID, chatID, streamID)

	messageObjID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		s.handleError(ctx, chatID, err)
		return
	}

	// Create cancellable context from the background context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		s.handleError(ctx, chatID, err)
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		s.handleError(ctx, chatID, err)
		return
	}

	// Store cancel function
	s.processesMu.Lock()
	s.activeProcesses[streamID] = cancel
	s.processesMu.Unlock()

	// Cleanup when done
	defer func() {
		s.processesMu.Lock()
		delete(s.activeProcesses, streamID)
		s.processesMu.Unlock()
	}()

	// Send initial processing message
	s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
		Event: "ai-response-step",
		Data:  "NeoBase is analyzing your request..",
	})

	// Get DB connection from dbManager
	dbConn, err := s.dbManager.GetConnection(chatID)
	if err != nil {
		s.handleError(ctx, chatID, err)
		return
	}

	// Get connection info
	connInfo, exists := s.dbManager.GetConnectionInfo(chatID)
	if !exists {
		s.handleError(ctx, chatID, fmt.Errorf("connection info not found"))
		return
	}

	// Fetch all the messages from the LLM
	messages, err := s.llmRepo.GetByChatID(chatObjID)
	if err != nil {
		s.handleError(ctx, chatID, err)
		return
	}

	// Helper function to check cancellation
	checkCancellation := func() bool {
		select {
		case <-ctx.Done():
			s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
				Event: "response-cancelled",
				Data:  "Operation cancelled by user",
			})
			return true
		default:
			return false
		}
	}

	// Check cancellation before expensive operations
	if checkCancellation() {
		return
	}

	// Check schema changes
	schemaManager := s.dbManager.GetSchemaManager()
	changed, err := schemaManager.HasSchemaChanged(ctx, chatID, dbConn)
	if err != nil {
		s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
			Event: "error",
			Data:  map[string]string{"error": err.Error()},
		})
		return
	}

	if checkCancellation() {
		return
	}

	if changed {
		// Only do full schema comparison if changes detected
		diff, err := schemaManager.CheckSchemaChanges(ctx, chatID, dbConn, connInfo.Config.Type)
		if err != nil {
			log.Printf("Error getting schema diff: %v", err)
		}

		if diff != nil {
			// Schema changed, get updated schema for LLM
			schema, err := schemaManager.GetSchema(ctx, chatID, dbConn, connInfo.Config.Type)
			if err != nil {
				log.Printf("Error getting schema: %v", err)
			} else {
				// Need to save this schema message to the LLM messages
				llmMsg := &models.LLMMessage{
					Base:      models.NewBase(),
					ChatID:    chatObjID,
					MessageID: messageObjID,
					Role:      string(MessageTypeSystem),
					Content: map[string]interface{}{
						"schema_update": fmt.Sprintf("%s\n\nChanges:\n%s",
							schemaManager.FormatSchemaForLLM(schema),
							s.formatSchemaDiff(diff),
						),
					},
				}
				if err := s.llmRepo.CreateMessage(llmMsg); err != nil {
					log.Printf("processLLMResponse -> Error saving LLM message: %v", err)
				}
				messages = append(messages, llmMsg)
			}
		}
	}

	// Send initial processing message
	s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
		Event: "ai-response-step",
		Data:  "Generating an optimized query & example results for the request..",
	})
	if checkCancellation() {
		return
	}

	// Generate LLM response
	response, err := s.llmClient.GenerateResponse(ctx, messages, connInfo.Config.Type)
	if err != nil {
		s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
			Event: "ai-response-error",
			Data:  map[string]string{"error": err.Error()},
		})
		return
	}

	log.Printf("processLLMResponse -> response: %s", response)

	if checkCancellation() {
		return
	}

	// Send initial processing message
	s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
		Event: "ai-response-step",
		Data:  "Analyzing the criticality of the query & if roll back is possible..",
	})

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
			Event: "ai-response-error",
			Data:  map[string]string{"error": err.Error()},
		})
	}

	queries := []models.Query{}
	for _, query := range jsonResponse["queries"].([]interface{}) {
		queryMap := query.(map[string]interface{})
		var exampleResult *string
		log.Printf("processLLMResponse -> queryMap: %v", queryMap)
		if queryMap["exampleResult"] != nil {
			log.Printf("processLLMResponse -> queryMap[\"exampleResult\"]: %v", queryMap["exampleResult"])
			result, _ := json.Marshal(queryMap["exampleResult"].([]interface{}))
			exampleResult = utils.ToStringPtr(string(result))
			log.Printf("processLLMResponse -> saving exampleResult: %v", *exampleResult)
		} else {
			exampleResult = nil
			log.Println("processLLMResponse -> saving exampleResult: nil")
		}

		var rollbackDependentQuery *string
		if queryMap["rollbackDependentQuery"] != nil {
			rollbackDependentQuery = utils.ToStringPtr(queryMap["rollbackDependentQuery"].(string))
		} else {
			rollbackDependentQuery = nil
		}
		queries = append(queries, models.Query{
			ID:                     primitive.NewObjectID(),
			Query:                  queryMap["query"].(string),
			Description:            queryMap["explanation"].(string),
			ExecutionTime:          nil,
			ExampleExecutionTime:   int(queryMap["estimateResponseTime"].(float64)),
			CanRollback:            queryMap["canRollback"].(bool),
			IsCritical:             queryMap["isCritical"].(bool),
			IsExecuted:             false,
			IsRolledBack:           false,
			ExampleResult:          exampleResult,
			ExecutionResult:        nil,
			Error:                  nil,
			QueryType:              utils.ToStringPtr(queryMap["queryType"].(string)),
			Tables:                 utils.ToStringPtr(queryMap["tables"].(string)),
			RollbackQuery:          utils.ToStringPtr(queryMap["rollbackQuery"].(string)),
			RollbackDependentQuery: rollbackDependentQuery,
		})
	}
	// Save response and send final message
	chatResponseMsg := models.NewMessage(userObjID, chatObjID, string(MessageTypeAssistant), jsonResponse["assistantMessage"].(string), &queries)
	if err := s.chatRepo.CreateMessage(chatResponseMsg); err != nil {
		s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
			Event: "error",
			Data:  map[string]string{"error": err.Error()},
		})
		return
	}

	formattedJsonResponse := map[string]interface{}{
		"assistant_response": jsonResponse,
	}
	llmMsg := &models.LLMMessage{
		Base:      models.NewBase(),
		ChatID:    chatObjID,
		MessageID: chatResponseMsg.ID,
		Content:   formattedJsonResponse,
		Role:      string(MessageTypeAssistant),
	}
	if err := s.llmRepo.CreateMessage(llmMsg); err != nil {
		log.Printf("processLLMResponse -> Error saving LLM message: %v", err)
	}

	// Send final response
	s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
		Event: "ai-response",
		Data: &dtos.MessageResponse{
			ID:        chatResponseMsg.ID.Hex(),
			ChatID:    chatResponseMsg.ChatID.Hex(),
			Content:   chatResponseMsg.Content,
			Queries:   dtos.ToQueryDto(chatResponseMsg.Queries),
			Type:      chatResponseMsg.Type,
			CreatedAt: chatResponseMsg.CreatedAt.Format(time.RFC3339),
		},
	})
}

// Rename formatSchemaUpdate to formatSchemaDiff and update its signature
func (s *chatService) formatSchemaDiff(diff *dbmanager.SchemaDiff) string {
	var msg strings.Builder
	msg.WriteString("Database schema has been updated:\n")

	if len(diff.AddedTables) > 0 {
		msg.WriteString("\nNew tables:\n")
		for _, t := range diff.AddedTables {
			msg.WriteString("- " + t + "\n")
		}
	}

	if len(diff.RemovedTables) > 0 {
		msg.WriteString("\nRemoved tables:\n")
		for _, t := range diff.RemovedTables {
			msg.WriteString("- " + t + "\n")
		}
	}

	if len(diff.ModifiedTables) > 0 {
		msg.WriteString("\nModified tables:\n")
		for table, changes := range diff.ModifiedTables {
			msg.WriteString(fmt.Sprintf("- %s:\n", table))
			if len(changes.AddedColumns) > 0 {
				msg.WriteString("  Added columns: " + strings.Join(changes.AddedColumns, ", ") + "\n")
			}
			if len(changes.RemovedColumns) > 0 {
				msg.WriteString("  Removed columns: " + strings.Join(changes.RemovedColumns, ", ") + "\n")
			}
			if len(changes.ModifiedColumns) > 0 {
				msg.WriteString("  Modified columns: " + strings.Join(changes.ModifiedColumns, ", ") + "\n")
			}
		}
	}

	return msg.String()
}

func (s *chatService) handleError(_ context.Context, chatID string, err error) {
	log.Printf("Error processing message for chat %s: %v", chatID, err)
}

func (s *chatService) UpdateMessage(userID, chatID, messageID string, req *dtos.CreateMessageRequest) (*dtos.MessageResponse, uint32, error) {
	// Implementation similar to CreateMessage but with update logic
	return nil, http.StatusNotImplemented, fmt.Errorf("not implemented")
}

func (s *chatService) DeleteMessages(userID, chatID string) (uint32, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid user ID format")
	}

	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid chat ID format")
	}

	// Verify chat ownership
	chat, err := s.chatRepo.FindByID(chatObjID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to fetch chat: %v", err)
	}
	if chat == nil {
		return http.StatusNotFound, fmt.Errorf("chat not found")
	}
	if chat.UserID != userObjID {
		return http.StatusForbidden, fmt.Errorf("unauthorized access to chat")
	}

	if err := s.chatRepo.DeleteMessages(chatObjID); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete messages: %v", err)
	}

	return http.StatusOK, nil
}

func (s *chatService) ListMessages(userID, chatID string, page, pageSize int) (*dtos.MessageListResponse, uint32, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid user ID format")
	}

	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid chat ID format")
	}

	// Verify chat ownership
	chat, err := s.chatRepo.FindByID(chatObjID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch chat: %v", err)
	}
	if chat == nil {
		return nil, http.StatusNotFound, fmt.Errorf("chat not found")
	}
	if chat.UserID != userObjID {
		return nil, http.StatusForbidden, fmt.Errorf("unauthorized access to chat")
	}

	messages, total, err := s.chatRepo.FindMessagesByChat(chatObjID, page, pageSize)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch messages: %v", err)
	}

	response := &dtos.MessageListResponse{
		Messages: make([]dtos.MessageResponse, len(messages)),
		Total:    total,
	}

	for i, msg := range messages {
		response.Messages[i] = *s.buildMessageResponse(msg)
	}

	return response, http.StatusOK, nil
}

// Helper methods for building responses
func (s *chatService) buildChatResponse(chat *models.Chat) *dtos.ChatResponse {
	return &dtos.ChatResponse{
		ID:     chat.ID.Hex(),
		UserID: chat.UserID.Hex(),
		Connection: dtos.ConnectionResponse{
			ID:       chat.ID.Hex(),
			Type:     chat.Connection.Type,
			Host:     chat.Connection.Host,
			Port:     chat.Connection.Port,
			Username: *chat.Connection.Username,
			Database: chat.Connection.Database,
		},
		CreatedAt: chat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: chat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *chatService) buildMessageResponse(msg *models.Message) *dtos.MessageResponse {
	return &dtos.MessageResponse{
		ID:        msg.ID.Hex(),
		ChatID:    msg.ChatID.Hex(),
		Type:      msg.Type,
		Content:   msg.Content,
		Queries:   dtos.ToQueryDto(msg.Queries),
		CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *chatService) SetStreamHandler(handler StreamHandler) {
	s.streamHandler = handler
}

// Helper method to send stream events
func (s *chatService) sendStreamEvent(userID, chatID, streamID string, response dtos.StreamResponse) {
	log.Printf("sendStreamEvent -> userID: %s, chatID: %s, streamID: %s, response: %+v", userID, chatID, streamID, response)
	if s.streamHandler != nil {
		s.streamHandler.HandleStreamEvent(userID, chatID, streamID, response)
	}
}

// Add method to cancel processing
func (s *chatService) CancelProcessing(streamID string) {
	s.processesMu.Lock()
	defer s.processesMu.Unlock()

	if cancel, exists := s.activeProcesses[streamID]; exists {
		cancel() // Cancel the context
		delete(s.activeProcesses, streamID)
	}
}

func (s *chatService) ConnectDB(ctx context.Context, userID, chatID string, streamID string) (uint32, error) {
	chatObjID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid chat ID format")
	}

	chatDetails, err := s.chatRepo.FindByID(chatObjID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to fetch chat: %v", err)
	}

	log.Printf("ChatService -> ConnectDB -> chatDetails: %+v", chatDetails)

	// Validate connection type
	if !isValidDBType(chatDetails.Connection.Type) {
		return http.StatusBadRequest, fmt.Errorf("unsupported database type")
	}

	// Subscribe to connection status updates before connecting
	s.dbManager.Subscribe(chatID, streamID)

	// Attempt to connect
	if err := s.dbManager.Connect(chatID, userID, streamID, dbmanager.ConnectionConfig{
		Type:     chatDetails.Connection.Type,
		Host:     chatDetails.Connection.Host,
		Port:     chatDetails.Connection.Port,
		Username: chatDetails.Connection.Username,
		Password: chatDetails.Connection.Password,
		Database: chatDetails.Connection.Database,
	}); err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to connect: %v", err)
	}

	log.Printf("ChatService -> ConnectDB -> connected to chat: %s", chatID)

	// Send to stream handler
	s.sendStreamEvent(userID, chatID, streamID, dtos.StreamResponse{
		Event: "db-connected",
		Data:  "Database connected successfully",
	})

	return http.StatusOK, nil
}

// Add method to handle DB status events
func (s *chatService) HandleDBEvent(userID, chatID, streamID string, response dtos.StreamResponse) {
	// Send to stream handler
	if s.streamHandler != nil {
		s.streamHandler.HandleStreamEvent(userID, chatID, streamID, response)
	}
}

func (s *chatService) DisconnectDB(ctx context.Context, userID, chatID string, streamID string) (uint32, error) {
	log.Printf("ChatService -> DisconnectDB -> Starting for chatID: %s", chatID)

	// Subscribe to connection status updates before disconnecting
	s.dbManager.Subscribe(chatID, streamID)
	log.Printf("ChatService -> DisconnectDB -> Subscribed to updates with streamID: %s", streamID)

	if err := s.dbManager.Disconnect(chatID, userID); err != nil {
		log.Printf("ChatService -> DisconnectDB -> failed to disconnect: %v", err)
		return http.StatusBadRequest, fmt.Errorf("failed to disconnect: %v", err)
	}

	log.Printf("ChatService -> DisconnectDB -> disconnected from chat: %s", chatID)
	return http.StatusOK, nil
}

func (s *chatService) GetDBConnectionStatus(ctx context.Context, userID, chatID string) (*dtos.ConnectionStatusResponse, uint32, error) {
	// Get connection info
	connInfo, exists := s.dbManager.GetConnectionInfo(chatID)
	if !exists {
		return nil, http.StatusNotFound, fmt.Errorf("no connection found")
	}

	// Check if connection is active
	isConnected := s.dbManager.IsConnected(chatID)

	// Convert port string to int
	port, err := strconv.Atoi(connInfo.Config.Port)
	if err != nil {
		port = 0 // Default value if conversion fails
	}

	return &dtos.ConnectionStatusResponse{
		IsConnected: isConnected,
		Type:        connInfo.Config.Type,
		Host:        connInfo.Config.Host,
		Port:        port,
		Database:    connInfo.Config.Database,
		Username:    *connInfo.Config.Username,
	}, http.StatusOK, nil
}

func isValidDBType(dbType string) bool {
	validTypes := []string{constants.DatabaseTypePostgreSQL, constants.DatabaseTypeMySQL, constants.DatabaseTypeMongoDB} // Add more as supported
	for _, t := range validTypes {
		if t == dbType {
			return true
		}
	}
	return false
}
