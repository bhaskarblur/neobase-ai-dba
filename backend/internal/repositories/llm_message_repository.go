package repositories

import (
	"context"
	"neobase-ai/internal/models"
	"neobase-ai/pkg/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LLMMessageRepository interface {
	// Message operations
	CreateMessage(message *models.LLMMessage) error
	UpdateMessage(id primitive.ObjectID, message *models.LLMMessage) error
	FindMessageByID(id primitive.ObjectID) (*models.LLMMessage, error)
	FindMessagesByChatID(chatID primitive.ObjectID) ([]*models.LLMMessage, int64, error)
	DeleteMessagesByChatID(chatID primitive.ObjectID) error
}

type llmMessageRepository struct {
	messageCollection *mongo.Collection
	streamCollection  *mongo.Collection
}

func NewLLMMessageRepository(mongoClient *mongodb.MongoDBClient) LLMMessageRepository {
	return &llmMessageRepository{
		messageCollection: mongoClient.GetCollectionByName("llm_messages"),
		streamCollection:  mongoClient.GetCollectionByName("llm_message_streams"),
	}
}

// Message operations
func (r *llmMessageRepository) CreateMessage(message *models.LLMMessage) error {
	_, err := r.messageCollection.InsertOne(context.Background(), message)
	return err
}

func (r *llmMessageRepository) UpdateMessage(id primitive.ObjectID, message *models.LLMMessage) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": message}
	_, err := r.messageCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *llmMessageRepository) FindMessageByID(id primitive.ObjectID) (*models.LLMMessage, error) {
	var message models.LLMMessage
	err := r.messageCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&message)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &message, err
}

func (r *llmMessageRepository) FindMessagesByChatID(chatID primitive.ObjectID) ([]*models.LLMMessage, int64, error) {
	var messages []*models.LLMMessage
	filter := bson.M{"chat_id": chatID}

	// Get total count
	total, err := r.messageCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.messageCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &messages)
	return messages, total, err
}

func (r *llmMessageRepository) DeleteMessagesByChatID(chatID primitive.ObjectID) error {
	filter := bson.M{"chat_id": chatID}
	_, err := r.messageCollection.DeleteMany(context.Background(), filter)
	return err

}
