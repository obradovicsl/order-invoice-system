package order

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"order-service/internal/repository"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

// MockOrderRepository implements Repository interface for testing
type MockOrderRepository struct {
	createOrderFunc       func(ctx context.Context, arg repository.CreateOrderParams) (repository.Order, error)
	createOrderItemFunc   func(ctx context.Context, arg repository.CreateOrderItemParams) (repository.OrderItem, error)
	getAllOrdersFunc      func(ctx context.Context) ([]repository.GetAllOrdersRow, error)
	getItemByIDFunc       func(ctx context.Context, id pgtype.UUID) (repository.GetItemByIDRow, error)
	getOrderByIDFunc      func(ctx context.Context, id pgtype.UUID) ([]repository.GetOrderByIDRow, error)
	updateItemStockFunc   func(ctx context.Context, arg repository.UpdateItemStockParams) (repository.Item, error)
	deleteOrderByIdFunc   func(ctx context.Context, id pgtype.UUID) error
	updateOrderStatusFunc func(ctx context.Context, arg repository.UpdateOrderStatusParams) (repository.Order, error)
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, arg repository.CreateOrderParams) (repository.Order, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, arg)
	}
	return repository.Order{}, errors.New("not implemented")
}

func (m *MockOrderRepository) CreateOrderItem(ctx context.Context, arg repository.CreateOrderItemParams) (repository.OrderItem, error) {
	if m.createOrderItemFunc != nil {
		return m.createOrderItemFunc(ctx, arg)
	}
	return repository.OrderItem{}, errors.New("not implemented")
}

func (m *MockOrderRepository) GetAllOrders(ctx context.Context) ([]repository.GetAllOrdersRow, error) {
	if m.getAllOrdersFunc != nil {
		return m.getAllOrdersFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockOrderRepository) GetItemByID(ctx context.Context, id pgtype.UUID) (repository.GetItemByIDRow, error) {
	if m.getItemByIDFunc != nil {
		return m.getItemByIDFunc(ctx, id)
	}
	return repository.GetItemByIDRow{}, errors.New("not implemented")
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, id pgtype.UUID) ([]repository.GetOrderByIDRow, error) {
	if m.getOrderByIDFunc != nil {
		return m.getOrderByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockOrderRepository) UpdateItemStock(ctx context.Context, arg repository.UpdateItemStockParams) (repository.Item, error) {
	if m.updateItemStockFunc != nil {
		return m.updateItemStockFunc(ctx, arg)
	}
	return repository.Item{}, errors.New("not implemented")
}

func (m *MockOrderRepository) DeleteOrderById(ctx context.Context, id pgtype.UUID) error {
	if m.deleteOrderByIdFunc != nil {
		return m.deleteOrderByIdFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *MockOrderRepository) UpdateOrderStatus(ctx context.Context, arg repository.UpdateOrderStatusParams) (repository.Order, error) {
	if m.updateOrderStatusFunc != nil {
		return m.updateOrderStatusFunc(ctx, arg)
	}
	return repository.Order{}, errors.New("not implemented")
}

// MockQueueService for testing
type MockQueueService struct {
	sendMessageFunc func(ctx context.Context, queueName string, messageText string, ttl int32) (string, error)
}

func (m *MockQueueService) SendMessage(ctx context.Context, queueName string, messageText string, ttl int32) (string, error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, queueName, messageText, ttl)
	}
	return "message-id", nil
}

// MockBlobService for testing
type MockBlobService struct {
	generateSignedURLFunc func(container, fileName string, expiry interface{}) (string, error)
}

func (m *MockBlobService) GenerateSignedURL(container, fileName string, expiry interface{}) (string, error) {
	if m.generateSignedURLFunc != nil {
		return m.generateSignedURLFunc(container, fileName, expiry)
	}
	return "signed-url", nil
}

func TestDeleteOrderById_InvalidUUID(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewService(mockRepo, logger, nil, "", nil, "")

	// Act
	err := svc.DeleteOrderById(context.Background(), "invalid-uuid")

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
	if err.Error() != "invalid order ID" {
		t.Errorf("expected 'invalid order ID', got '%s'", err.Error())
	}
}

func TestDeleteOrderById_Success(t *testing.T) {
	// Arrange
	testUUID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

	mockRepo := &MockOrderRepository{
		deleteOrderByIdFunc: func(ctx context.Context, id pgtype.UUID) error {
			if id != testUUID {
				t.Errorf("expected UUID to match, got different value")
			}
			return nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewService(mockRepo, logger, nil, "", nil, "")

	// Act
	err := svc.DeleteOrderById(context.Background(), testUUID.String())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewService(mockRepo, logger, nil, "", nil, "")

	userID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
	req := CreateOrderRequest{
		UserID:   userID,
		UserName: "test-user",
		Items:    []CreateOrderItemRequest{}, // Empty items
	}

	// Act
	_, err := svc.CreateOrder(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty items, got nil")
	}
	if err.Error() != "order must have at least one item" {
		t.Errorf("expected 'order must have at least one item', got '%s'", err.Error())
	}
}

func TestCreateOrder_InvalidQuantity(t *testing.T) {
	// Arrange
	mockRepo := &MockOrderRepository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewService(mockRepo, logger, nil, "", nil, "")

	userID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
	itemID := pgtype.UUID{Bytes: [16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

	req := CreateOrderRequest{
		UserID:   userID,
		UserName: "test-user",
		Items: []CreateOrderItemRequest{
			{
				ItemID:   itemID,
				Quantity: 0, // Invalid quantity
			},
		},
	}

	// Act
	_, err := svc.CreateOrder(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid quantity, got nil")
	}
	if err.Error() != "quantity must be greater than 0" {
		t.Errorf("expected 'quantity must be greater than 0', got '%s'", err.Error())
	}
}
