package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

type BlobService struct {
	client           *azblob.Client
	containerName    string
	logger           *slog.Logger
	connectionString string
}

type UploadBlobInput struct {
	ContainerName string
	BlobName      string
	Data          []byte
}

type DownloadBlobOutput struct {
	Data []byte
}

func NewBlobService(connectionString string, containerName string, logger *slog.Logger) (*BlobService, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		logger.Error("failed to create blob client", "error", err)
		return nil, err
	}

	return &BlobService{
		client:           client,
		containerName:    containerName,
		logger:           logger,
		connectionString: connectionString,
	}, nil
}

// extractAccountKeyFromConnectionString extracts the account key from a connection string
func extractAccountKeyFromConnectionString(connectionString string) string {
	parts := strings.Split(connectionString, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "AccountKey=") {
			return strings.TrimPrefix(part, "AccountKey=")
		}
	}
	return ""
}

// extractAccountNameFromConnectionString extracts the account name from a connection string
func extractAccountNameFromConnectionString(connectionString string) string {
	parts := strings.Split(connectionString, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "AccountName=") {
			return strings.TrimPrefix(part, "AccountName=")
		}
	}
	return ""
}

// UploadBlob uploads data to Azure Blob Storage
func (bs *BlobService) UploadBlob(ctx context.Context, input UploadBlobInput) error {
	bs.logger.Info("uploading blob to Azure",
		"container", input.ContainerName,
		"blob_name", input.BlobName,
		"data_size", len(input.Data),
	)

	_, err := bs.client.UploadBuffer(ctx, input.ContainerName, input.BlobName, input.Data, nil)
	if err != nil {
		bs.logger.Error("failed to upload blob", "blob_name", input.BlobName, "error", err)
		return fmt.Errorf("failed to upload blob: %w", err)
	}

	bs.logger.Info("blob uploaded successfully", "blob_name", input.BlobName)
	return nil
}

// DownloadBlob downloads data from Azure Blob Storage
func (bs *BlobService) DownloadBlob(ctx context.Context, containerName string, blobName string) (*DownloadBlobOutput, error) {
	bs.logger.Info("downloading blob from Azure",
		"container", containerName,
		"blob_name", blobName,
	)

	downloadResponse, err := bs.client.DownloadStream(ctx, containerName, blobName, nil)
	if err != nil {
		bs.logger.Error("failed to download blob", "blob_name", blobName, "error", err)
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}

	data, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		bs.logger.Error("failed to read blob data", "blob_name", blobName, "error", err)
		return nil, fmt.Errorf("failed to read blob data: %w", err)
	}

	defer downloadResponse.Body.Close()

	bs.logger.Info("blob downloaded successfully",
		"blob_name", blobName,
		"data_size", len(data),
	)

	return &DownloadBlobOutput{
		Data: data,
	}, nil
}

// DeleteBlob deletes a blob from Azure Blob Storage
func (bs *BlobService) DeleteBlob(ctx context.Context, containerName string, blobName string) error {
	bs.logger.Info("deleting blob from Azure",
		"container", containerName,
		"blob_name", blobName,
	)

	_, err := bs.client.DeleteBlob(ctx, containerName, blobName, nil)
	if err != nil {
		bs.logger.Error("failed to delete blob", "blob_name", blobName, "error", err)
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	bs.logger.Info("blob deleted successfully", "blob_name", blobName)
	return nil
}

// CreateContainer creates a container if it doesn't exist
func (bs *BlobService) CreateContainer(ctx context.Context, containerName string) error {
	bs.logger.Info("creating container", "container", containerName)

	_, err := bs.client.CreateContainer(ctx, containerName, &container.CreateOptions{})
	if err != nil {
		// Container might already exist, check if it's a conflict error
		bs.logger.Error("failed to create container", "container", containerName, "error", err)
		return fmt.Errorf("failed to create container: %w", err)
	}

	bs.logger.Info("container created successfully", "container", containerName)
	return nil
}

// ListBlobs lists all blobs in a container
func (bs *BlobService) ListBlobs(ctx context.Context, containerName string) ([]string, error) {
	bs.logger.Info("listing blobs", "container", containerName)

	pager := bs.client.NewListBlobsFlatPager(containerName, nil)

	var blobNames []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			bs.logger.Error("failed to list blobs", "container", containerName, "error", err)
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range page.Segment.BlobItems {
			blobNames = append(blobNames, *blob.Name)
		}
	}

	bs.logger.Info("blobs listed successfully",
		"container", containerName,
		"blob_count", len(blobNames),
	)

	return blobNames, nil
}

// GenerateSignedURL generates a SAS (Shared Access Signature) URL for direct blob access
func (bs *BlobService) GenerateSignedURL(containerName string, blobName string, expiryDuration time.Duration) (string, error) {
	bs.logger.Info("generating signed URL for blob",
		"container", containerName,
		"blob_name", blobName,
		"expiry_duration", expiryDuration.String(),
	)

	// Extract account name and key from connection string
	accountName := extractAccountNameFromConnectionString(bs.connectionString)
	accountKey := extractAccountKeyFromConnectionString(bs.connectionString)

	if accountName == "" || accountKey == "" {
		bs.logger.Error("failed to extract account name or key from connection string")
		return "", fmt.Errorf("invalid connection string: missing AccountName or AccountKey")
	}

	// Create SAS signature values
	sv := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     time.Now().UTC().Add(-time.Minute),
		ExpiryTime:    time.Now().UTC().Add(expiryDuration),
		Permissions:   "r",
		ContainerName: containerName,
		BlobName:      blobName,
	}

	// Sign with shared key (using manual HMAC-SHA256 signing)
	signature := computeHMACSHA256(sv, accountKey, accountName)
	sasToken := fmt.Sprintf("sv=%s&se=%s&sr=b&sp=%s&sig=%s",
		url.QueryEscape("2021-06-08"),
		url.QueryEscape(sv.ExpiryTime.UTC().Format("2006-01-02T15:04:05Z")),
		url.QueryEscape(sv.Permissions),
		url.QueryEscape(signature),
	)

	// Construct the signed URL
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s", accountName, containerName, blobName, sasToken)

	bs.logger.Info("signed URL generated successfully", "blob_name", blobName)
	return blobURL, nil
}

// computeHMACSHA256 computes HMAC-SHA256 signature for SAS token
func computeHMACSHA256(sv sas.BlobSignatureValues, accountKey string, accountName string) string {
	// Decode the base64 account key
	decodedKey, _ := base64.StdEncoding.DecodeString(accountKey)

	// Create the string to sign following Azure SAS token format
	// StringToSign = signedPermissions + "\n" +
	//                signedStart + "\n" +
	//                signedExpiry + "\n" +
	//                canonicalizedResource + "\n" +
	//                signedIdentifier + "\n" +
	//                signedIP + "\n" +
	//                signedProtocol + "\n" +
	//                signedVersion
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s\n\n\n\n%s",
		sv.Permissions,
		sv.StartTime.UTC().Format("2006-01-02T15:04:05Z"),
		sv.ExpiryTime.UTC().Format("2006-01-02T15:04:05Z"),
		fmt.Sprintf("/blob/%s/%s/%s", accountName, sv.ContainerName, sv.BlobName),
		"2021-06-08",
	)

	// Compute HMAC-SHA256
	h := hmac.New(sha256.New, decodedKey)
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}
