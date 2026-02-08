package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
)

type QueueService struct {
	connectionString string
	logger           *slog.Logger
}

type SendMessageInput struct {
	QueueName   string
	MessageText string
	TTL         int32 // Time to live in seconds (optional, default is 604800 - 7 days)
}

type QueueMessage struct {
	MessageID       string
	PopReceipt      string
	MessageText     string
	DequeueCount    int64
	ExpirationTime  string
	InsertionTime   string
	TimeNextVisible string
}

func NewQueueService(connectionString string, logger *slog.Logger) (*QueueService, error) {
	return &QueueService{
		connectionString: connectionString,
		logger:           logger,
	}, nil
}

// SendMessage sends a message to Azure Queue Storage
func (qs *QueueService) SendMessage(ctx context.Context, input SendMessageInput) (string, error) {
	qs.logger.Info("sending message to queue",
		"queue", input.QueueName,
		"message_length", len(input.MessageText),
	)

	queueClient, err := azqueue.NewQueueClientFromConnectionString(
		qs.connectionString,
		input.QueueName,
		nil,
	)
	if err != nil {
		qs.logger.Error("failed to create queue client", "queue", input.QueueName, "error", err)
		return "", err
	}

	sendOptions := &azqueue.EnqueueMessageOptions{}
	if input.TTL > 0 {
		sendOptions.TimeToLive = &input.TTL
	}

	response, err := queueClient.EnqueueMessage(ctx, input.MessageText, sendOptions)
	if err != nil {
		qs.logger.Error("failed to send message", "queue", input.QueueName, "error", err)
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	if len(response.Messages) == 0 {
		return "", fmt.Errorf("no message returned from enqueue")
	}

	messageID := *response.Messages[0].MessageID

	qs.logger.Info("message sent successfully",
		"queue", input.QueueName,
		"message_id", messageID,
	)

	return messageID, nil
}

// DequeueMessage dequeues a message from Azure Queue Storage
func (qs *QueueService) DequeueMessage(ctx context.Context, queueName string, visibilityTimeout int32) (*QueueMessage, error) {
	// qs.logger.Info("dequeuing message from queue",
	// 	"queue", queueName,
	// 	"visibility_timeout", visibilityTimeout,
	// )

	queueClient, err := azqueue.NewQueueClientFromConnectionString(
		qs.connectionString,
		queueName,
		nil,
	)
	if err != nil {
		qs.logger.Error("failed to create queue client", "queue", queueName, "error", err)
		return nil, err
	}

	dequeueOptions := &azqueue.DequeueMessageOptions{}
	if visibilityTimeout > 0 {
		dequeueOptions.VisibilityTimeout = &visibilityTimeout
	}

	response, err := queueClient.DequeueMessage(ctx, dequeueOptions)
	if err != nil {
		qs.logger.Error("failed to dequeue message", "queue", queueName, "error", err)
		return nil, fmt.Errorf("failed to dequeue message: %w", err)
	}

	// If no messages available
	if len(response.Messages) == 0 {
		// qs.logger.Info("no messages in queue", "queue", queueName)
		return nil, nil
	}

	msg := response.Messages[0]

	return &QueueMessage{
		MessageID:       *msg.MessageID,
		PopReceipt:      *msg.PopReceipt,
		MessageText:     *msg.MessageText,
		DequeueCount:    *msg.DequeueCount,
		ExpirationTime:  msg.ExpirationTime.String(),
		InsertionTime:   msg.InsertionTime.String(),
		TimeNextVisible: msg.TimeNextVisible.String(),
	}, nil
}

// DeleteMessage deletes a message from Azure Queue Storage
func (qs *QueueService) DeleteMessage(ctx context.Context, queueName string, messageID string, popReceipt string) error {
	qs.logger.Info("deleting message from queue",
		"queue", queueName,
		"message_id", messageID,
	)

	queueClient, err := azqueue.NewQueueClientFromConnectionString(
		qs.connectionString,
		queueName,
		nil,
	)
	if err != nil {
		qs.logger.Error("failed to create queue client", "queue", queueName, "error", err)
		return err
	}

	_, err = queueClient.DeleteMessage(ctx, messageID, popReceipt, nil)
	if err != nil {
		qs.logger.Error("failed to delete message", "message_id", messageID, "error", err)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	qs.logger.Info("message deleted successfully", "message_id", messageID)
	return nil
}

// CreateQueue creates a queue if it doesn't exist
func (qs *QueueService) CreateQueue(ctx context.Context, queueName string) error {
	qs.logger.Info("creating queue", "queue", queueName)

	queueClient, err := azqueue.NewQueueClientFromConnectionString(
		qs.connectionString,
		queueName,
		nil,
	)
	if err != nil {
		qs.logger.Error("failed to create queue client", "queue", queueName, "error", err)
		return err
	}

	_, err = queueClient.Create(ctx, nil)
	if err != nil {
		qs.logger.Error("failed to create queue", "queue", queueName, "error", err)
		return fmt.Errorf("failed to create queue: %w", err)
	}

	qs.logger.Info("queue created successfully", "queue", queueName)
	return nil
}

// DeleteQueue deletes a queue
func (qs *QueueService) DeleteQueue(ctx context.Context, queueName string) error {
	qs.logger.Info("deleting queue", "queue", queueName)

	queueClient, err := azqueue.NewQueueClientFromConnectionString(
		qs.connectionString,
		queueName,
		nil,
	)
	if err != nil {
		qs.logger.Error("failed to create queue client", "queue", queueName, "error", err)
		return err
	}

	_, err = queueClient.Delete(ctx, nil)
	if err != nil {
		qs.logger.Error("failed to delete queue", "queue", queueName, "error", err)
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	qs.logger.Info("queue deleted successfully", "queue", queueName)
	return nil
}

// PeekMessage peeks at a message without removing it from the queue
func (qs *QueueService) PeekMessage(ctx context.Context, queueName string) (*QueueMessage, error) {
	qs.logger.Info("peeking message from queue", "queue", queueName)

	queueClient, err := azqueue.NewQueueClientFromConnectionString(
		qs.connectionString,
		queueName,
		nil,
	)
	if err != nil {
		qs.logger.Error("failed to create queue client", "queue", queueName, "error", err)
		return nil, err
	}

	response, err := queueClient.PeekMessage(ctx, nil)
	if err != nil {
		qs.logger.Error("failed to peek message", "queue", queueName, "error", err)
		return nil, fmt.Errorf("failed to peek message: %w", err)
	}

	if len(response.Messages) == 0 {
		qs.logger.Info("no messages to peek", "queue", queueName)
		return nil, nil
	}

	msg := response.Messages[0]

	return &QueueMessage{
		MessageID:       *msg.MessageID,
		PopReceipt:      "",
		MessageText:     *msg.MessageText,
		DequeueCount:    *msg.DequeueCount,
		ExpirationTime:  msg.ExpirationTime.String(),
		InsertionTime:   msg.InsertionTime.String(),
		TimeNextVisible: "",
	}, nil
}

// SendMessageWithEncoding sends a Base64 encoded message to the queue
func (qs *QueueService) SendMessageWithEncoding(ctx context.Context, input SendMessageInput) (string, error) {
	qs.logger.Info("sending encoded message to queue",
		"queue", input.QueueName,
		"message_length", len(input.MessageText),
	)

	encodedMessage := base64.StdEncoding.EncodeToString([]byte(input.MessageText))

	return qs.SendMessage(ctx, SendMessageInput{
		QueueName:   input.QueueName,
		MessageText: encodedMessage,
		TTL:         input.TTL,
	})
}
