package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// InvoiceMessage represents the complete order data message from queue
type InvoiceMessage struct {
	OrderID    string          `json:"order_id"`
	UserID     string          `json:"user_id"`
	UserName   string          `json:"user_name"`
	OrderPrice interface{}     `json:"order_price"`
	Items      []interface{}   `json:"items"`
	Timestamp  int64           `json:"timestamp"`
	MessageID  string          `json:"message_id"`
}

type InvoiceWorker struct {
	queueService    *QueueService
	blobService     *BlobService
	invoiceService  *InvoiceService
	queueName       string
	readyQueueName  string
	logger          *slog.Logger
}

func NewInvoiceWorker(
	queueService *QueueService,
	blobService *BlobService,
	invoiceService *InvoiceService,
	queueName string,
	readyQueueName string,
	logger *slog.Logger,
) *InvoiceWorker {
	return &InvoiceWorker{
		queueService:    queueService,
		blobService:     blobService,
		invoiceService:  invoiceService,
		queueName:       queueName,
		readyQueueName:  readyQueueName,
		logger:          logger,
	}
}

// Start begins listening to the queue
func (w *InvoiceWorker) Start(ctx context.Context, pollInterval time.Duration) error {
	w.logger.Info("starting invoice worker",
		"queue", w.queueName,
		"poll_interval", pollInterval.String(),
	)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("invoice worker received shutdown signal")
			return ctx.Err()
		case <-ticker.C:
			w.processMessages(ctx)
		}
	}
}

// processMessages polls the queue and processes available messages
func (w *InvoiceWorker) processMessages(ctx context.Context) {
	message, err := w.queueService.DequeueMessage(ctx, w.queueName, 30)
	if err != nil {
		w.logger.Error("failed to dequeue message", "error", err)
		return
	}

	if message == nil {
		// No messages available
		return
	}

	w.logger.Info("processing message", "message_id", message.MessageID)

	// Parse message
	invoiceMsg, err := w.parseMessage(message.MessageText)
	if err != nil {
		w.logger.Error("failed to parse message", "message_id", message.MessageID, "error", err)
		// Delete the invalid message
		if delErr := w.queueService.DeleteMessage(ctx, w.queueName, message.MessageID, message.PopReceipt); delErr != nil {
			w.logger.Error("failed to delete invalid message", "message_id", message.MessageID, "error", delErr)
		}
		return
	}

	// Process the invoice
	if err := w.generateAndUploadInvoice(ctx, invoiceMsg); err != nil {
		w.logger.Error("failed to process invoice",
			"order_id", invoiceMsg.OrderID,
			"message_id", message.MessageID,
			"error", err,
		)
		// Don't delete on error - let it retry
		return
	}

	// Send completion message to ready queue
	w.logger.Info("sending completion message to ready queue", "order_id", invoiceMsg.OrderID)
	if err := w.sendCompletionMessage(ctx, invoiceMsg.OrderID); err != nil {
		w.logger.Error("failed to send completion message",
			"order_id", invoiceMsg.OrderID,
			"error", err,
		)
		// Don't delete on error - let it retry
		return
	}

	// Delete the message after successful processing
	if err := w.queueService.DeleteMessage(ctx, w.queueName, message.MessageID, message.PopReceipt); err != nil {
		w.logger.Error("failed to delete processed message",
			"message_id", message.MessageID,
			"error", err,
		)
		return
	}

	w.logger.Info("message processed successfully",
		"message_id", message.MessageID,
		"order_id", invoiceMsg.OrderID,
	)
}

// parseMessage parses and validates the queue message
func (w *InvoiceWorker) parseMessage(messageText string) (*InvoiceMessage, error) {
	var msg InvoiceMessage
	if err := json.Unmarshal([]byte(messageText), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Validate message format
	if msg.OrderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}

	if msg.Timestamp == 0 {
		return nil, fmt.Errorf("timestamp is required")
	}

	if msg.MessageID == "" {
		return nil, fmt.Errorf("message_id is required")
	}

	// Validate message_id format (should be order_id+timestamp)
	expectedMessageID := fmt.Sprintf("%s_%d", msg.OrderID, msg.Timestamp)
	if msg.MessageID != expectedMessageID {
		return nil, fmt.Errorf("invalid message_id: expected %s, got %s", expectedMessageID, msg.MessageID)
	}

	return &msg, nil
}

// messageToOrderResponse converts InvoiceMessage to OrderResponse for PDF generation
func (w *InvoiceWorker) messageToOrderResponse(msg *InvoiceMessage) (*OrderResponse, error) {
	// Parse order ID to UUID
	var orderID pgtype.UUID
	if err := orderID.Scan(msg.OrderID); err != nil {
		w.logger.Warn("could not parse order ID as UUID", "order_id", msg.OrderID)
	}

	// Parse user ID to UUID
	var userID pgtype.UUID
	if err := userID.Scan(msg.UserID); err != nil {
		w.logger.Warn("could not parse user ID as UUID", "user_id", msg.UserID)
	}

	// Convert items from interface{} to OrderItem
	items := make([]OrderItem, 0)
	for _, itemData := range msg.Items {
		itemMap, ok := itemData.(map[string]interface{})
		if !ok {
			w.logger.Warn("could not convert item data", "item", itemData)
			continue
		}

		item := OrderItem{
			ItemCode: getString(itemMap, "code"),
			ItemName: getString(itemMap, "name"),
			Quantity: int32(getFloat(itemMap, "quantity")),
		}
		items = append(items, item)
	}

	// Create order price as pgtype.Numeric
	// For now, just create empty numeric - invoice service should handle this
	orderPrice := pgtype.Numeric{}

	return &OrderResponse{
		ID:         orderID,
		UserID:     userID,
		UserName:   msg.UserName,
		OrderPrice: orderPrice,
		Status:     "PENDING",
		Items:      items,
	}, nil
}

// Helper functions to extract values from interface{} maps
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}

// generateAndUploadInvoice generates PDF from message data and uploads to blob
func (w *InvoiceWorker) generateAndUploadInvoice(ctx context.Context, msg *InvoiceMessage) error {
	w.logger.Info("generating invoice for order", "order_id", msg.OrderID)

	// Convert message data to OrderResponse
	orderResp, err := w.messageToOrderResponse(msg)
	if err != nil {
		return fmt.Errorf("failed to convert message data: %w", err)
	}

	// Generate PDF
	pdfData, err := w.invoiceService.GenerateInvoicePDF(orderResp)
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Generate filename
	filename := w.invoiceService.InvoiceFileName(msg.OrderID)

	// Upload to blob storage
	uploadInput := UploadBlobInput{
		ContainerName: w.blobService.containerName,
		BlobName:      filename,
		Data:          pdfData,
	}

	if err := w.blobService.UploadBlob(ctx, uploadInput); err != nil {
		return fmt.Errorf("failed to upload invoice to blob: %w", err)
	}

	w.logger.Info("invoice generated and uploaded successfully",
		"order_id", msg.OrderID,
		"filename", filename,
		"size", len(pdfData),
	)

	return nil
}

// sendCompletionMessage sends a completion message to the ready queue
func (w *InvoiceWorker) sendCompletionMessage(ctx context.Context, orderID string) error {
	timestamp := time.Now().Unix()
	messageID := fmt.Sprintf("completed_%s_%d", orderID, timestamp)

	message := map[string]interface{}{
		"order_id":   orderID,
		"timestamp":  timestamp,
		"message_id": messageID,
	}

	messageText, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal completion message: %w", err)
	}

	_, err = w.queueService.SendMessage(ctx, w.readyQueueName, string(messageText), 0)
	if err != nil {
		return fmt.Errorf("failed to send completion message: %w", err)
	}

	w.logger.Info("completion message sent to ready queue",
		"order_id", orderID,
		"message_id", messageID,
	)

	return nil
}

// ValidateQueueAndContainer ensures queue and blob container exist
func (w *InvoiceWorker) ValidateQueueAndContainer(ctx context.Context, containerName string) error {
	w.logger.Info("validating queue and container",
		"queue", w.queueName,
		"container", containerName,
	)

	// Create queue if it doesn't exist
	if err := w.queueService.CreateQueue(ctx, w.queueName); err != nil {
		w.logger.Warn("queue creation warning",
			"queue", w.queueName,
			"error", err,
		)
		// Don't fail - queue might already exist
	}

	// Create container if it doesn't exist
	if err := w.blobService.CreateContainer(ctx, containerName); err != nil {
		w.logger.Warn("container creation warning",
			"container", containerName,
			"error", err,
		)
		// Don't fail - container might already exist
	}

	w.logger.Info("queue and container validation completed")
	return nil
}
