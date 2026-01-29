package service

import (
	"context"
	"encoding/json"
	"fmt"
	"invoice-worker/internal/repository"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InvoiceMessage represents the message format in the queue
type InvoiceMessage struct {
	OrderID   string `json:"order_id"`
	Timestamp int64  `json:"timestamp"`
	MessageID string `json:"message_id"`
}

type InvoiceWorker struct {
	queueService   *QueueService
	blobService    *BlobService
	invoiceService *InvoiceService
	orderService   *OrderService
	dbPool         *pgxpool.Pool
	queueName      string
	logger         *slog.Logger
}

func NewInvoiceWorker(
	queueService *QueueService,
	blobService *BlobService,
	invoiceService *InvoiceService,
	orderService *OrderService,
	dbPool *pgxpool.Pool,
	queueName string,
	logger *slog.Logger,
) *InvoiceWorker {
	return &InvoiceWorker{
		queueService:   queueService,
		blobService:    blobService,
		invoiceService: invoiceService,
		orderService:   orderService,
		dbPool:         dbPool,
		queueName:      queueName,
		logger:         logger,
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

	// Update order status to PROCESSING
	w.logger.Info("updating order status to PROCESSING", "order_id", invoiceMsg.OrderID)
	if err := w.orderService.UpdateOrderStatus(ctx, invoiceMsg.OrderID, repository.OrderStatusPROCESSING); err != nil {
		w.logger.Error("failed to update order status to PROCESSING",
			"order_id", invoiceMsg.OrderID,
			"error", err,
		)
		// Don't delete on error - let it retry
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

	// Update order status to COMPLETED
	w.logger.Info("updating order status to COMPLETED", "order_id", invoiceMsg.OrderID)
	if err := w.orderService.UpdateOrderStatus(ctx, invoiceMsg.OrderID, repository.OrderStatusCOMPLETED); err != nil {
		w.logger.Error("failed to update order status to COMPLETED",
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

// generateAndUploadInvoice retrieves order data, generates PDF, and uploads to blob
func (w *InvoiceWorker) generateAndUploadInvoice(ctx context.Context, msg *InvoiceMessage) error {
	w.logger.Info("generating invoice for order", "order_id", msg.OrderID)

	// Fetch order from database
	orderResp, err := w.orderService.GetOrderByID(ctx, msg.OrderID)
	if err != nil {
		return fmt.Errorf("failed to fetch order: %w", err)
	}

	if orderResp == nil {
		return fmt.Errorf("order not found")
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
