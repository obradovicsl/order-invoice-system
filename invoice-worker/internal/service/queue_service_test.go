package service

import (
	"io"
	"log/slog"
	"testing"
)

func TestNewQueueService(t *testing.T) {
	// Arrange
	connectionString := "DefaultEndpointsProtocol=https;AccountName=test;AccountKey=key=="
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Act
	queueService, err := NewQueueService(connectionString, logger)

	// Assert
	if err != nil {
		t.Fatalf("expected no error during NewQueueService, got %v", err)
	}

	if queueService == nil {
		t.Fatal("expected non-nil QueueService, got nil")
	}

	if queueService.connectionString != connectionString {
		t.Errorf("expected connectionString to be set, got %q", queueService.connectionString)
	}

	if queueService.logger == nil {
		t.Fatal("expected non-nil logger, got nil")
	}
}

func TestSendMessageInput_Validation(t *testing.T) {
	// Arrange
	input := SendMessageInput{
		QueueName:   "test-queue",
		MessageText: "test message",
		TTL:         3600,
	}

	// Act & Assert
	if input.QueueName == "" {
		t.Error("expected non-empty QueueName")
	}

	if input.MessageText == "" {
		t.Error("expected non-empty MessageText")
	}

	if input.TTL != 3600 {
		t.Errorf("expected TTL=3600, got %d", input.TTL)
	}
}

func TestQueueMessageStruct(t *testing.T) {
	// Arrange
	queueMsg := QueueMessage{
		MessageID:       "msg-123",
		PopReceipt:      "receipt-123",
		MessageText:     "test message content",
		DequeueCount:    1,
		ExpirationTime:  "2024-02-10T12:00:00Z",
		InsertionTime:   "2024-02-10T11:00:00Z",
		TimeNextVisible: "2024-02-10T11:05:00Z",
	}

	// Act & Assert
	if queueMsg.MessageID != "msg-123" {
		t.Errorf("expected MessageID='msg-123', got '%s'", queueMsg.MessageID)
	}

	if queueMsg.PopReceipt != "receipt-123" {
		t.Errorf("expected PopReceipt='receipt-123', got '%s'", queueMsg.PopReceipt)
	}

	if queueMsg.MessageText != "test message content" {
		t.Errorf("expected MessageText='test message content', got '%s'", queueMsg.MessageText)
	}

	if queueMsg.DequeueCount != 1 {
		t.Errorf("expected DequeueCount=1, got %d", queueMsg.DequeueCount)
	}
}

func TestSendMessageInputWithoutTTL(t *testing.T) {
	// Arrange
	input := SendMessageInput{
		QueueName:   "default-queue",
		MessageText: "message without TTL",
		TTL:         0, // Default TTL (7 days)
	}

	// Act & Assert
	if input.TTL != 0 {
		t.Errorf("expected TTL=0 (default), got %d", input.TTL)
	}

	if input.QueueName != "default-queue" {
		t.Errorf("expected QueueName='default-queue', got '%s'", input.QueueName)
	}
}
