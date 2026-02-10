package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"order-service/internal/repository"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CompletionMessage represents the completion message from invoice worker
type CompletionMessage struct {
	OrderID   string `json:"order_id"`
	Timestamp int64  `json:"timestamp"`
	MessageID string `json:"message_id"`
}

type OrderCompletionWorker struct {
	queueService   *QueueService
	queries        *repository.Queries
	dbPool         *pgxpool.Pool
	readyQueueName string
	logger         *slog.Logger
}

func NewOrderCompletionWorker(
	queueService *QueueService,
	queries *repository.Queries,
	dbPool *pgxpool.Pool,
	readyQueueName string,
	logger *slog.Logger,
) *OrderCompletionWorker {
	return &OrderCompletionWorker{
		queueService:   queueService,
		queries:        queries,
		dbPool:         dbPool,
		readyQueueName: readyQueueName,
		logger:         logger,
	}
}

// Start begins listening to the ready queue
func (w *OrderCompletionWorker) Start(ctx context.Context, pollInterval time.Duration) error {
	w.logger.Info("starting order completion worker",
		"queue", w.readyQueueName,
		"poll_interval", pollInterval.String(),
	)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("order completion worker received shutdown signal")
			return ctx.Err()
		case <-ticker.C:
			w.processMessages(ctx)
		}
	}
}

// processMessages polls the ready queue and processes available messages
func (w *OrderCompletionWorker) processMessages(ctx context.Context) {
	message, err := w.queueService.DequeueMessage(ctx, w.readyQueueName, 30)
	if err != nil {
		w.logger.Error("failed to dequeue message", "error", err)
		return
	}

	if message == nil {
		// No messages available
		return
	}

	w.logger.Info("processing completion message", "message_id", message.MessageID)

	// Parse message
	completionMsg, err := w.parseMessage(message.MessageText)
	if err != nil {
		w.logger.Error("failed to parse completion message", "message_id", message.MessageID, "error", err)
		// Delete the invalid message
		if delErr := w.queueService.DeleteMessage(ctx, w.readyQueueName, message.MessageID, message.PopReceipt); delErr != nil {
			w.logger.Error("failed to delete invalid message", "message_id", message.MessageID, "error", delErr)
		}
		return
	}

	// Update order status to COMPLETED
	w.logger.Info("updating order status to COMPLETED", "order_id", completionMsg.OrderID)
	if err := w.updateOrderStatus(ctx, completionMsg.OrderID); err != nil {
		w.logger.Error("failed to update order status to COMPLETED",
			"order_id", completionMsg.OrderID,
			"error", err,
		)
		// Don't delete on error - let it retry
		return
	}

	// Delete the message after successful processing
	if err := w.queueService.DeleteMessage(ctx, w.readyQueueName, message.MessageID, message.PopReceipt); err != nil {
		w.logger.Error("failed to delete processed message",
			"message_id", message.MessageID,
			"error", err,
		)
		return
	}

	w.logger.Info("completion message processed successfully",
		"message_id", message.MessageID,
		"order_id", completionMsg.OrderID,
	)
}

// parseMessage parses and validates the completion message
func (w *OrderCompletionWorker) parseMessage(messageText string) (*CompletionMessage, error) {
	var msg CompletionMessage
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

	return &msg, nil
}

// updateOrderStatus updates the order status to COMPLETED
func (w *OrderCompletionWorker) updateOrderStatus(ctx context.Context, orderID string) error {
	// Parse order ID string to pgtype.UUID
	var uuid pgtype.UUID
	if err := uuid.Scan(orderID); err != nil {
		return fmt.Errorf("invalid order ID format: %w", err)
	}

	// Execute update in transaction
	tx, err := w.dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := w.queries.WithTx(tx)
	_, err = qtx.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
		ID:     uuid,
		Status: repository.OrderStatusCOMPLETED,
	})
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	w.logger.Info("order status updated to COMPLETED", "order_id", orderID)
	return nil
}

// ValidateQueue ensures the ready queue exists
func (w *OrderCompletionWorker) ValidateQueue(ctx context.Context) error {
	w.logger.Info("validating ready queue", "queue", w.readyQueueName)

	// Create queue if it doesn't exist
	if err := w.queueService.CreateQueue(ctx, w.readyQueueName); err != nil {
		w.logger.Warn("queue creation warning",
			"queue", w.readyQueueName,
			"error", err,
		)
		// Don't fail - queue might already exist
	}

	w.logger.Info("ready queue validation completed")
	return nil
}
