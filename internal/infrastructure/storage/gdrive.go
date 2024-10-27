package storage

import (
	"context"
	"mime/multipart"

	"github.com/dk5761/go-serv/configs"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GDriveStorageService struct {
	service  *drive.Service
	folderID string
}

func NewGDriveStorageService(cfg configs.GDriveConfig) (*GDriveStorageService, error) {
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(cfg.CredentialsJSON))
	if err != nil {
		return nil, err
	}
	return &GDriveStorageService{
		service:  srv,
		folderID: cfg.FolderID,
	}, nil
}

func (s *GDriveStorageService) UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error) {
	defer file.Close()
	f := &drive.File{
		Name:    fileName,
		Parents: []string{s.folderID},
	}

	res, err := s.service.Files.Create(f).Media(file).Do()
	if err != nil {
		return "", err
	}
	return res.Id, nil
}
