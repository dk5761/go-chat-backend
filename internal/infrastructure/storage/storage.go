package storage

import (
	"context"
	"mime/multipart"
)

type StorageService interface {
	UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error)
	// Add other methods if needed
}
