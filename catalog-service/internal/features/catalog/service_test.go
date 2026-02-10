package catalog

import (
	"catalog-service/internal/repository"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// MockRepository implements Repository interface for testing
type MockRepository struct {
	createCatalogItemFunc  func(ctx context.Context, arg repository.CreateCatalogItemParams) (repository.Item, error)
	getAllCatalogItemsFunc func(ctx context.Context) ([]repository.Item, error)
	getCatalogItemByIDFunc func(ctx context.Context, id pgtype.UUID) (repository.Item, error)
	updateItemQuantityFunc func(ctx context.Context, arg repository.UpdateItemQuantityParams) (repository.Item, error)
}

func (m *MockRepository) CreateCatalogItem(ctx context.Context, arg repository.CreateCatalogItemParams) (repository.Item, error) {
	if m.createCatalogItemFunc != nil {
		return m.createCatalogItemFunc(ctx, arg)
	}
	return repository.Item{}, errors.New("not implemented")
}

func (m *MockRepository) GetAllCatalogItems(ctx context.Context) ([]repository.Item, error) {
	if m.getAllCatalogItemsFunc != nil {
		return m.getAllCatalogItemsFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetCatalogItemByID(ctx context.Context, id pgtype.UUID) (repository.Item, error) {
	if m.getCatalogItemByIDFunc != nil {
		return m.getCatalogItemByIDFunc(ctx, id)
	}
	return repository.Item{}, errors.New("not implemented")
}

func (m *MockRepository) UpdateItemQuantity(ctx context.Context, arg repository.UpdateItemQuantityParams) (repository.Item, error) {
	if m.updateItemQuantityFunc != nil {
		return m.updateItemQuantityFunc(ctx, arg)
	}
	return repository.Item{}, errors.New("not implemented")
}

// MockBlobService for testing - implements the methods needed by Service
type MockBlobService struct {
	generateUploadSignedURLFunc func(container, fileName string, expiry time.Duration) (string, string, error)
}

func (m *MockBlobService) GenerateUploadSignedURL(container, fileName string, expiry time.Duration) (string, string, error) {
	if m.generateUploadSignedURLFunc != nil {
		return m.generateUploadSignedURLFunc(container, fileName, expiry)
	}
	return "signed-url", "download-url", nil
}

// Dummy implementations for other BlobService methods so it can be used as *service.BlobService in tests
func (m *MockBlobService) UploadBlob(ctx context.Context, input interface{}) error { return nil }
func (m *MockBlobService) DownloadBlob(ctx context.Context, containerName, blobName string) (interface{}, error) {
	return nil, nil
}
func (m *MockBlobService) DeleteBlob(ctx context.Context, containerName, blobName string) error {
	return nil
}
func (m *MockBlobService) CreateContainer(ctx context.Context, containerName string) error {
	return nil
}
func (m *MockBlobService) ListBlobs(ctx context.Context, containerName string) ([]string, error) {
	return nil, nil
}
func (m *MockBlobService) GenerateSignedURL(containerName, blobName string, expiryDuration time.Duration) (string, error) {
	return "", nil
}

func TestGetAllCatalogItems_Success(t *testing.T) {
	// Arrange
	mockItems := []repository.Item{
		{
			Code: "ITEM001",
			Name: "Test Item 1",
		},
		{
			Code: "ITEM002",
			Name: "Test Item 2",
		},
	}

	mockRepo := &MockRepository{
		getAllCatalogItemsFunc: func(ctx context.Context) ([]repository.Item, error) {
			return mockItems, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewService(mockRepo, logger, nil, "")

	// Act
	items, err := svc.GetAllCatalogItems(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if items[0].Code != "ITEM001" {
		t.Errorf("expected first item code 'ITEM001', got '%s'", items[0].Code)
	}
}

func TestGetCatalogItemByID_Success(t *testing.T) {
	// Arrange
	testID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
	expectedItem := repository.Item{
		ID:   testID,
		Code: "ITEM123",
		Name: "Test Item",
	}

	mockRepo := &MockRepository{
		getCatalogItemByIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Item, error) {
			if id != testID {
				t.Errorf("expected ID to match")
			}
			return expectedItem, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewService(mockRepo, logger, nil, "")

	// Act
	item, err := svc.GetCatalogItemByID(context.Background(), testID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.Code != "ITEM123" {
		t.Errorf("expected item code 'ITEM123', got '%s'", item.Code)
	}
}
