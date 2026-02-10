// +build integration

package main

import (
	"bytes"
	"catalog-service/internal/config"
	"catalog-service/internal/features/catalog"
	"catalog-service/internal/features/catalog/handler"
	"catalog-service/internal/logger"
	"catalog-service/internal/repository"
	"catalog-service/internal/server"
	"catalog-service/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testPool *pgxpool.Pool
)

func setupPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "catalog_test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, "", err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, "", err
	}

	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/catalog_test", host, port.Port())
	return container, connStr, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start Postgres container
	container, connStr, err := setupPostgresContainer(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to start Postgres container: %v", err))
	}
	defer container.Terminate(ctx)

	// Create pool
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse connection string: %v", err))
	}
	testPool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create pool: %v", err))
	}
	defer testPool.Close()

	// Run migrations
	if err := runTestMigrations(ctx, testPool); err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}

	// Run tests
	code := m.Run()
	exit(code)
}

func runTestMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			price NUMERIC(10, 2) NOT NULL,
			stock_quantity INTEGER NOT NULL DEFAULT 0,
			image_url VARCHAR(255),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return err
		}
	}

	return nil
}

// Test 1: Create Catalog Item
func TestCreateCatalogItem(t *testing.T) {

	// Initialize services with test pool
	log := logger.NewLogger(logger.LoggerConfig{DefaultLevel: "error"})
	queries := repository.New(testPool)
	blobService, _ := service.NewBlobService(
		"DefaultEndpointsProtocol=https;AccountName=test;AccountKey=test==;EndpointSuffix=core.windows.net",
		"test-container",
		log,
	)

	catalogService := catalog.NewService(queries, log, blobService, "test-container")

	cfg := &config.Config{
		AllowedOrigins: []string{"*"},
		Logger:         logger.LoggerConfig{DefaultLevel: "error"},
	}
	router := server.NewRouter(*cfg, catalogService, log)

	// Create test request
	createItemRequest := handler.CreateCatalogItemRequest{
		Code:          "TEST-ITEM-" + fmt.Sprintf("%d", time.Now().Unix()),
		Name:          "Test Item",
		Price:         99.99,
		StockQuantity: 10,
		ImageUrl:      nil,
	}

	body, _ := json.Marshal(createItemRequest)
	req := httptest.NewRequest("POST", "/api/v1/catalog/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Errorf("Expected status 200 or 201, got %d. Response: %s", w.Code, w.Body.String())
	}

	var response handler.ItemResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != createItemRequest.Code {
		t.Errorf("Expected code %s, got %s", createItemRequest.Code, response.Code)
	}

	t.Logf("✓ Test 1 Passed: Create Catalog Item - ID: %s", response.ID)
}

// Test 2: Get Catalog Item by ID
func TestGetCatalogItemByID(t *testing.T) {

	// Initialize services with test pool
	log := logger.NewLogger(logger.LoggerConfig{DefaultLevel: "error"})
	queries := repository.New(testPool)
	blobService, _ := service.NewBlobService(
		"DefaultEndpointsProtocol=https;AccountName=test;AccountKey=test==;EndpointSuffix=core.windows.net",
		"test-container",
		log,
	)

	catalogService := catalog.NewService(queries, log, blobService, "test-container")

	cfg := &config.Config{
		AllowedOrigins: []string{"*"},
		Logger:         logger.LoggerConfig{DefaultLevel: "error"},
	}
	router := server.NewRouter(*cfg, catalogService, log)

	// First, create an item
	createItemRequest := handler.CreateCatalogItemRequest{
		Code:          "TEST-GET-ITEM-" + fmt.Sprintf("%d", time.Now().Unix()),
		Name:          "Get Test Item",
		Price:         49.99,
		StockQuantity: 5,
		ImageUrl:      nil,
	}

	body, _ := json.Marshal(createItemRequest)
	createReq := httptest.NewRequest("POST", "/api/v1/catalog/", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()

	router.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated && createW.Code != http.StatusOK {
		t.Fatalf("Failed to create item for get test: %d", createW.Code)
	}

	var createdItem handler.ItemResponse
	json.NewDecoder(createW.Body).Decode(&createdItem)

	// Now, get the item by ID
	getReq := httptest.NewRequest("GET", "/api/v1/catalog/"+createdItem.ID, nil)
	getW := httptest.NewRecorder()

	router.ServeHTTP(getW, getReq)

	// Assert
	if getW.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", getW.Code, getW.Body.String())
	}

	var retrievedItem handler.ItemResponse
	json.NewDecoder(getW.Body).Decode(&retrievedItem)

	if retrievedItem.ID != createdItem.ID {
		t.Errorf("Expected ID %s, got %s", createdItem.ID, retrievedItem.ID)
	}

	t.Logf("✓ Test 2 Passed: Get Catalog Item by ID - %s", retrievedItem.ID)
}

func exit(code int) {
	// Exit code is handled by testing framework
}
